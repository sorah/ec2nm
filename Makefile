Gomfile.lock: Gomfile
	gondler install

install: Gomfile.lock build

build:
	gondler build -o bin/ec2nm

run: build
	./bin/ec2nm
