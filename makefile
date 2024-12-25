generate:
	protoc --proto_path=proto proto/*.proto --go_out=. --go-grpc_out=.

commands:
	grpcurl -d @ -proto proto\movie_proto.proto -import-path ./ -plaintext localhost:50051 moviebase.UserService/CreateUsers < create_users.json 
	# getting single user with grpcurl and cmd
	
	grpcurl -d "{\"id\": \"d77ef8ba-c2b1-11ef-900a-54ee756d8952\"}" -proto proto\movie_proto.proto -import-path ./ -plaintext localhost:50051 moviebase.UserService/GetUser


