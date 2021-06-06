---
version: '3'
services:
  satis:
    image: blueacornici/satis:satis
    restart: always
    privileged: true
    network_mode: bridge
    environment:
      - VIRTUAL_HOST={{.domain_name}}
      - LETSENCRYPT_HOST={{.domain_name}}
      - LETSENCRYPT_EMAIL=devops+ssl@blueacorn.com
      - HOMEPAGE={{.domain_name}}
    expose:
      - 80
    volumes:
      - /docker-volumes/satisfy/composer:/var/www/.composer/cache/:rw
      - /docker-volumes/satisfy/dist:/satisfy/web/dist:rw
