include /usr/local/IBFD/etc/docker.env
include version.env

PROJECT=docsan
FILES=-f docker-compose.yml -f local-compose.yml

silentstart:
	@docker-compose ${FILES} pull 2>/dev/null >/dev/null && docker-compose ${FILES} -p${PROJECT} up -d 2>/dev/null >/dev/null

silentstop:
	@docker-compose ${FILES} -p${PROJECT} down 2>/dev/null >/dev/null

start:
	docker-compose ${FILES} -p${PROJECT} up -d

stop:
	docker-compose ${FILES} -p${PROJECT} down

refresh:
	docker-compose ${FILES} pull 

