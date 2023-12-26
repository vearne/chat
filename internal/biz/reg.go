package biz

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	zlog "github.com/vearne/chat/internal/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.uber.org/zap"
	"path"
	"time"
)

// ServiceRegister 创建租约注册服务
type ServiceRegister struct {
	cli     *clientv3.Client //etcd client
	leaseID clientv3.LeaseID //租约ID
	//租约keepalieve相应chan
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string //key
	val           string //value
}

// NewServiceRegister 新建注册服务
func NewServiceRegister(
	etcdServers []string,
	username, password string,
	prefix, addr string,
) (*ServiceRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Username:    username,
		Password:    password,
		Endpoints:   etcdServers,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		zlog.Fatal("NewServiceRegister", zap.Error(err))
	}

	buff, _ := json.Marshal(endpoints.Endpoint{Addr: addr})
	key := path.Join(prefix, uuid.NewString())
	zlog.Info("register service", zap.String("key", key), zap.String("val", string(buff)))
	ser := &ServiceRegister{
		cli: cli,
		key: key,
		val: string(buff),
	}

	//申请租约设置时间keepalive
	// 5 seconds
	var lease int64 = 5
	if err := ser.putKeyWithLease(lease); err != nil {
		return nil, err
	}

	return ser, nil
}

// 设置租约
func (s *ServiceRegister) putKeyWithLease(lease int64) error {
	//设置租约时间
	resp, err := s.cli.Grant(context.Background(), lease)
	if err != nil {
		return err
	}
	//注册服务并绑定租约
	_, err = s.cli.Put(context.Background(), s.key, s.val, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	//设置续租 定期发送需求请求
	leaseRespChan, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	zlog.Debug("putKeyWithLease", zap.Int64("leaseID", int64(s.leaseID)))
	s.keepAliveChan = leaseRespChan
	return nil
}

// ListenLeaseRespChan 监听 续租情况
func (s *ServiceRegister) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {
		zlog.Debug("lease renew successful", zap.Any("leaseKeepResp", leaseKeepResp))
	}
}

// Close 注销服务
func (s *ServiceRegister) Close() error {
	if _, err := s.cli.Revoke(context.Background(), s.leaseID); err != nil {
		return err
	}
	zlog.Info("cancel lease")
	return s.cli.Close()
}
