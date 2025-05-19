
## Bootstrap node setup

## Copy database required init files
- scp -i {path_to_rsa_private_key} source_path/{file} dest_path:/home/ec2-user/app/database/{file}

## Run the go build

For linux distribution, the GOOS and GOARCH should be setup.

- GOOS="linux" GOARCH="amd64" go build -o app . "from the root"

## Create a symlink on instance
- ln -s app/{path}/tbb usr/ec2-user/local/bin/tbb

## Run the bootstrap node
- tbb run --dir=./database --bootstrap --port=8080
