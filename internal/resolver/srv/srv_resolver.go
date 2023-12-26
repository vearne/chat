package srv

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
	"net"
	"strconv"
	"sync"
	"time"
)

var logger = grpclog.Component("dns")

var (
	errMissingAddr = errors.New("dns resolver: missing address")

	// Addresses ending with a colon that is supposed to be the separator
	// between host and port is not allowed.  E.g. "::" is a valid address as
	// it is an IPv6 address (host only) and "[::]:" is invalid as it ends with
	// a colon as the host and port separator
	errEndsWithColon = errors.New("dns resolver: missing port after port-separator colon")

	defaultResolver netResolver = net.DefaultResolver
	minDNSResRate               = 10 * time.Second
)

type netResolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
	LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
	LookupTXT(ctx context.Context, name string) (txts []string, err error)
}

func init() {
	resolver.Register(NewSRVBuilder())
}

func NewSRVBuilder() resolver.Builder {
	return &srvBuilder{}
}

type srvBuilder struct{}

// Build creates and starts a DNS resolver that watches the name resolution of the target.
func (b *srvBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	host, port, err := parseTarget(target.Endpoint(), "8080")
	if err != nil {
		return nil, err
	}

	// IP address.
	if ipAddr, ok := formatIP(host); ok {
		addr := []resolver.Address{{Addr: ipAddr + ":" + port}}
		err = cc.UpdateState(resolver.State{Addresses: addr})
		return deadResolver{}, err
	}

	// DNS address (non-IP).
	ctx, cancel := context.WithCancel(context.Background())
	d := &srvResolver{
		host:                 host,
		port:                 port,
		ctx:                  ctx,
		cancel:               cancel,
		cc:                   cc,
		rn:                   make(chan struct{}, 1),
		disableServiceConfig: opts.DisableServiceConfig,
	}

	d.resolver = defaultResolver
	d.wg.Add(1)
	go d.watcher()
	return d, nil
}

// Scheme returns the naming scheme of this resolver builder, which is "dns".
func (b *srvBuilder) Scheme() string {
	return "srv"
}

// dnsResolver watches for the name resolution update for a non-IP target.
type srvResolver struct {
	host     string
	port     string
	resolver netResolver
	ctx      context.Context
	cancel   context.CancelFunc
	cc       resolver.ClientConn
	// rn channel is used by ResolveNow() to force an immediate resolution of the target.
	rn chan struct{}
	// wg is used to enforce Close() to return after the watcher() goroutine has finished.
	// Otherwise, data race will be possible. [Race Example] in dns_resolver_test we
	// replace the real lookup functions with mocked ones to facilitate testing.
	// If Close() doesn't wait for watcher() goroutine finishes, race detector sometimes
	// will warns lookup (READ the lookup function pointers) inside watcher() goroutine
	// has data race with replaceNetFunc (WRITE the lookup function pointers).
	wg                   sync.WaitGroup
	disableServiceConfig bool
}

func (d *srvResolver) Close() {
	d.cancel()
	d.wg.Wait()
}

// ResolveNow invoke an immediate resolution of the target that this dnsResolver watches.
func (d *srvResolver) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case d.rn <- struct{}{}:
	default:
	}
}

func (d *srvResolver) watcher() {
	defer d.wg.Done()
	backoffIndex := 1
	for {
		state, err := d.lookup()
		if err != nil {
			// Report error to the underlying grpc.ClientConn.
			d.cc.ReportError(err)
		} else {
			err = d.cc.UpdateState(*state)
		}

		var timer *time.Timer
		if err == nil {
			// Success resolving, wait for the next ResolveNow. However, also wait 30 seconds at the very least
			// to prevent constantly re-resolving.
			backoffIndex = 1
			timer = time.NewTimer(minDNSResRate)
			select {
			case <-d.ctx.Done():
				timer.Stop()
				return
			case <-d.rn:
			}
		} else {
			timer = time.NewTimer(time.Second * 3)
			backoffIndex++
		}
		select {
		case <-d.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (d *srvResolver) lookup() (*resolver.State, error) {
	srvs, srvErr := d.lookupSRV()
	if srvErr != nil || len(srvs) == 0 {
		return nil, srvErr
	}

	addrs := make([]resolver.Address, 0, len(srvs))
	for _, srv := range srvs {
		addrs = append(addrs, resolver.Address{Addr: srv.Addr, ServerName: srv.ServerName})
	}
	state := resolver.State{Addresses: addrs}
	return &state, nil
}

func (d *srvResolver) lookupSRV() ([]resolver.Address, error) {
	var newAddrs []resolver.Address
	_, srvs, err := d.resolver.LookupSRV(d.ctx, "grpclb", "tcp", d.host)
	if err != nil {
		err = handleDNSError(err, "SRV") // may become nil
		return nil, err
	}
	for _, s := range srvs {
		lbAddrs, err := d.resolver.LookupHost(d.ctx, s.Target)
		if err != nil {
			err = handleDNSError(err, "A") // may become nil
			if err == nil {
				// If there are other SRV records, look them up and ignore this
				// one that does not exist.
				continue
			}
			return nil, err
		}
		for _, a := range lbAddrs {
			ip, ok := formatIP(a)
			if !ok {
				return nil, fmt.Errorf("dns: error parsing A record IP address %v", a)
			}
			addr := ip + ":" + strconv.Itoa(int(s.Port))
			newAddrs = append(newAddrs, resolver.Address{Addr: addr, ServerName: s.Target})
		}
	}
	return newAddrs, nil
}

func handleDNSError(err error, lookupType string) error {
	err = filterError(err)
	if err != nil {
		err = fmt.Errorf("dns: %v record lookup error: %v", lookupType, err)
		logger.Info(err)
	}
	return err
}

var filterError = func(err error) error {
	if dnsErr, ok := err.(*net.DNSError); ok && !dnsErr.IsTimeout && !dnsErr.IsTemporary {
		// Timeouts and temporary errors should be communicated to gRPC to
		// attempt another DNS query (with backoff).  Other errors should be
		// suppressed (they may represent the absence of a TXT record).
		return nil
	}
	return err
}

// deadResolver is a resolver that does nothing.
type deadResolver struct{}

func (deadResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (deadResolver) Close() {}

// formatIP returns ok = false if addr is not a valid textual representation of an IP address.
// If addr is an IPv4 address, return the addr and ok = true.
// If addr is an IPv6 address, return the addr enclosed in square brackets and ok = true.
func formatIP(addr string) (addrIP string, ok bool) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return "", false
	}
	if ip.To4() != nil {
		return addr, true
	}
	return "[" + addr + "]", true
}

// parseTarget takes the user input target string and default port, returns formatted host and port info.
// If target doesn't specify a port, set the port to be the defaultPort.
// If target is in IPv6 format and host-name is enclosed in square brackets, brackets
// are stripped when setting the host.
// examples:
// target: "www.google.com" defaultPort: "443" returns host: "www.google.com", port: "443"
// target: "ipv4-host:80" defaultPort: "443" returns host: "ipv4-host", port: "80"
// target: "[ipv6-host]" defaultPort: "443" returns host: "ipv6-host", port: "443"
// target: ":80" defaultPort: "443" returns host: "localhost", port: "80"
func parseTarget(target, defaultPort string) (host, port string, err error) {
	if target == "" {
		return "", "", errMissingAddr
	}
	if ip := net.ParseIP(target); ip != nil {
		// target is an IPv4 or IPv6(without brackets) address
		return target, defaultPort, nil
	}
	if host, port, err = net.SplitHostPort(target); err == nil {
		if port == "" {
			// If the port field is empty (target ends with colon), e.g. "[::1]:", this is an error.
			return "", "", errEndsWithColon
		}
		// target has port, i.e ipv4-host:port, [ipv6-host]:port, host-name:port
		if host == "" {
			// Keep consistent with net.Dial(): If the host is empty, as in ":80", the local system is assumed.
			host = "localhost"
		}
		return host, port, nil
	}
	if host, port, err = net.SplitHostPort(target + ":" + defaultPort); err == nil {
		// target doesn't have port
		return host, port, nil
	}
	return "", "", fmt.Errorf("invalid target address %v, error info: %v", target, err)
}
