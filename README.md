* Для сборки проекта используйте  (находясь в /thumbnail_utility)
  go build -o bin/server_preview server/server.go
  go build -o bin/client_preview client/client.go

* Запустите в отдельных вкладках терминала (находясь в /thumbnail_utility)
  bin/server_preview (запускает сервер на локальном хосте)

  bin/client_preview (посылает соответствующий аргументам командной строки запрос на сервер)
  
  В утилите bin/client_preview [--async] https://... 
ключ --async опциональный, при его наличии адреса будут обрабатываться асинхронно и после ключа следуют urls, разделенные пробелами вида:
https://www.youtube.com/watch?v=VideoId... или https://www.youtube.com/VideoId... 
  
   Пример команды утилиты: bin/client_preview --async  https://www.youtube.com/watch?v=7g01DlHlQqI https://www.youtube.com/watch?v=7g01DlHlQqI&list=PL4_hYwCyhAvYyx4TIRk6tLG0c8CLGzhE5&index=1&ab_channel=%D0%9B%D0%B5%D0%BA%D1%82%D0%BE%D1%80%D0%B8%D0%B9%D0%A4%D0%9F%D0%9C%D0%98
  
* Возможно потребуется поставить sqlite3, к примеру на Ubuntu введите две команды 

    sudo apt update
    sudo apt install sqlite3

