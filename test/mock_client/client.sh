# list
grpcurl  --plaintext 127.0.0.1:18223 list
grpcurl -import-path ../../logic -proto logic.proto list

# describe
grpcurl  --plaintext 127.0.0.1:18223 describe
grpcurl  --plaintext 127.0.0.1:18223 describe logic.LogicDealer

# create account
grpcurl --plaintext -d '{"nickname": "zhangsan","broker": "dev1:18080"}'\
 127.0.0.1:18223 logic.LogicDealer/CreateAccount

grpcurl --plaintext -d '{"nickname": "lisi","broker": "dev1:18080"}'\
 127.0.0.1:18223 logic.LogicDealer/CreateAccount

# match
grpcurl --plaintext -d '{"accountId":10}'\
 127.0.0.1:18223 logic.LogicDealer/Match
