go get github.com/ilyakaznacheev/cleanenv
go get modernc.org/sqlite
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5  


sqlite:/home/xybug/Downloads/golangchatapp/sqlite/dev/api.db

Add headers like:

![image](images/paste_1784133976.png)

Add message:

![image](images/paste_1784134062.png)


#### Serve the html
cd /path/to/your/html
python3 -m http.server 8000

Then open:
http://localhost:8000


Create a new folder outside the backend and make a `main.go`


Run: User A

go run main.go \
-url=ws://localhost:8082/api/ws \
-token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJuYW1lIjoiYWJheW9taSIsIlgtcGxhdGZvcm0iOiJ3ZWIiLCJzdWIiOiIxIiwiZXhwIjoxNzg0MjcwMjIwfQ.skTlZOJhNo02I8YxMtfJnP0lj9cuS28RlgZwKvnsIbU \
-name=Alice

Run: User B
go run main.go \
-token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJuYW1lIjoibWltaSIsIlgtcGxhdGZvcm0iOiJ3ZWIiLCJzdWIiOiIyIiwiZXhwIjoxNzg0MjcwMzQ1fQ.pDOZ9DSp_oz9N9EeLP7CH3ynSGcC_2gsbyaRQmrrvO0 \
-name=Bob


Send a message

Copy and paste this into Alice's terminal:
{
  "event_type": "message",
  "payload": {
    "private_id": 1,
    "receiver_id": 2,
    "message_type": "text",
    "content": "Hello Bob!"
  }
}

Bob should receive something like:
{
  "event_type": "message",
  "payload": {
    "message": {
      ...
    }
  }
}

