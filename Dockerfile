FROM golang:1.25-alpine

WORKDIR /app

#RUN apk add --no-cache \
#        fontconfig \
#        ttf-freefont \
#        libxrender \
#        libx11 \
#        libgcc \
#        libstdc++ \
#        qt5-qtwebsockets \
#        qt5-qtwebchannel \
#        qt5-qtwebengine \
#        qt5-qtsvg \
#        bash

#RUN wget https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6-1/wkhtmltox-0.12.6-1.alpine.pub20160324.tar.xz
#RUN tar -xf wkhtmltox-0.12.6-1.alpine.pub20160324.tar.xz \
#    mv wkhtmltox/bin/wkhtmltopdf /usr/local/bin/  \
#    mv wkhtmltox/bin/wkhtmltoimage /usr/local/bin/ \
#    chmod +x /usr/local/bin/wkhtmltopdf \

COPY . .

RUN go mod tidy

RUN go build -o bot ./internal

CMD ["./bot"]