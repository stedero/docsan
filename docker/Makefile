include /usr/local/IBFD/etc/docker.env
include version.env

PROJECT=docsan

silentstart:
	@docker-compose pull 2>/dev/null >/dev/null && docker-compose -p${PROJECT} up -d 2>/dev/null >/dev/null

silentstop:
	@docker-compose -p${PROJECT} down 2>/dev/null >/dev/null

start:
	docker-compose -p${PROJECT} up -d

stop:
	docker-compose -p${PROJECT} down

refresh:
	docker-compose pull 

