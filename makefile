syntetic:
	python utils/dataset_hashid.py

run:
	sudo rm -rf /tmp/*
	go run ./cmd/app/main.go serve