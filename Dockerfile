FROM python:3.13.5

WORKDIR /app

COPY ./requirements.txt /app/requirements.txt

RUN pip install -r /app/requirements.txt

COPY . /app

CMD [ "fastapi", "run", "api.py", "--port", "8000"]