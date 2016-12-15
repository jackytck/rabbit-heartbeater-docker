heartbeater: *.go
	go build -o bin/heartbeater

.env: .env
	cp .env bin

worker-linux: *.go
	env GOOS=linux GOARCH=amd64 go build -o bin/heartbeater

.env-linux: .env.prod
	cp .env.prod bin/.env

linux: worker-linux

main: heartbeater .env

publish: linux .env-linux
	rsync -azPv --delete bin/ rabbit-ec2:/home/altizure/alti-heartbeater/bin
	ssh rabbit-ec2 sudo systemctl restart alti-heartbeater

clean:
	rm -rf bin
