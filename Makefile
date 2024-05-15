up_build:
	@echo Stopping running containers...
	docker-compose down -v
	@echo Builing new images if required and starting images...
	docker-compose up --build -d
	@echo Docker images build and started !!! 
stop:
	@echo Stopping running containers...
	docker-compose down -v
	@echo Containers stopped and removed 