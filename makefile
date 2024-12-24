generate:
	protoc --proto_path=proto proto/*.proto --go_out=. --go-grpc_out=.

commands:
	grpcurl -d @ -proto proto\movie_proto.proto -import-path ./ -plaintext localhost:50051 moviebase.UserService/CreateUsers < create_users.json 
	# getting single user
	
	grpcurl -d "{\"id\": \"53b59d38-c215-11ef-972e-54ee756d8952\"}" -proto proto\movie_proto.proto -import-path ./ -plaintext localhost:50051 moviebase.UserService/GetUser


