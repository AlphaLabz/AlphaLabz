FROM alpine:3.18

WORKDIR /pb

RUN apk add --no-cache bash

COPY pocketbase /pb/pocketbase
COPY init-pocketbase.sh /pb/init-pocketbase.sh

RUN chmod +x /pb/init-pocketbase.sh

EXPOSE 8090

CMD ["sh", "/pb/init-pocketbase.sh"]