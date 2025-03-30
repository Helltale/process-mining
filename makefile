syntetic:
	python utils/dataset_hashid.py

run:
	sudo rm -rf /tmp/*
	go run ./cmd/app/main.go serve

sh:
	chmod +x run.sh
	./run.sh