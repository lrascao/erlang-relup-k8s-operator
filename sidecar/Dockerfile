FROM alpine:3.9

RUN apk add python3
RUN apk add py3-pip

RUN mkdir /app
WORKDIR /app
ADD . /app/
RUN pip3 install -r requirements.txt

CMD ["python3", "-u", "/app/main.py"]
