generate:
	protoc --proto_path=proto proto/*.proto --go_out=. --go-grpc_out=.

commands:
	grpcurl -d @ -proto proto\movie_proto.proto -import-path ./ -plaintext localhost:50051 moviebase.UserService/CreateUsers < create_users.json 
	# getting single user
	
	grpcurl -d "{\"id\": \"248f8388-c217-11ef-93ca-54ee756d8952\"}" -proto proto\movie_proto.proto -import-path ./ -plaintext localhost:50051 moviebase.UserService/GetUser


