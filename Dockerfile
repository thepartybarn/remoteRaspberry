FROM balenalib/raspberrypi3-ubuntu

RUN mkdir /app

WORKDIR /app

COPY remoteraspberry .

CMD ./remoteraspberry
