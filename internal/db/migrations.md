mkdir -p migrations
migrate create -ext sql -dir migrations -seq create_users_table
migrate create -ext sql -dir migrations -seq create_privates_table
migrate create -ext sql -dir migrations -seq create_messages_table
migrate create -ext sql -dir migrations -seq add_indexes

 
