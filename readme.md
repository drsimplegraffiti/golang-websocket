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
-token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJuYW1lIjoiSm9obiIsIlgtcGxhdGZvcm0iOiJ3ZWIiLCJzdWIiOiIzIiwiZXhwIjoxNzg0MTQyMjQ0fQ.KevXUWBvIVh8qa2g0rpan5IF159LfdhjB5M84LfKtIc \
-name=Alice

Run: User B
go run main.go \
-token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJuYW1lIjoiYWJheW9taSIsIlgtcGxhdGZvcm0iOiJ3ZWIiLCJzdWIiOiIxIiwiZXhwIjoxNzg0MTQyODI3fQ.qd6FLHd2oV_7Z7mMyIAY-iZCRQnD0QlZIYV6X8E5GhE \
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

