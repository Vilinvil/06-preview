* Для генерации кода на Go из api.proto выполнить, находясь в директории thumbnail_utility

    protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative  api/api.proto


* Чтобы поставить sqlite в Ubuntu, введите две команды 

    sudo apt update
    sudo apt install sqlite3