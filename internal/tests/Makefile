all:
	mkdir -p testpb
	docker run --platform=linux/amd64 -v $(PWD):/defs namely/protoc-all -f testpb.proto -l go
	mv gen/pb-go/github.com/chucklehead-dev/gosim/internal/tests/testpb/*.pb.go testpb/
	rm -rf gen
