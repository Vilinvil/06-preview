* Для сборки проекта используйте
  (находясь в /thumbnail_utility/server)
  go build -o ../bin/server_preview cmd/main.go
  (находясь в /thumbnail_utility/client)
  go build -o ../bin/client_preview cmd/main.go

* Запустите в отдельных вкладках терминала (находясь в /thumbnail_utility)
  bin/server_preview (запускает сервер на локальном хосте)

  bin/client_preview (посылает соответствующий аргументам командной строки запрос на сервер)
  
  В утилите bin/client_preview [--async] 'https://...' 
Ключ --async опциональный, при его наличии адреса будут обрабатываться асинхронно и после ключа следуют urls внутри одинарных кавычек, разделенные пробелами вида:
https://www.youtube.com/watch?v=VideoId... или https://www.youtube.com/VideoId... 
Можно использовать и без кавычек, но тогда в терминале нужно экранировать специальные символы
  
  Примеры команды утилиты:
bin/client_preview 'https://www.youtube.com/watch?v=jfKfPfyJRdk&ab_channel=LofiGirl https://www.youtube.com/watch?v=mesl2Si6saw https://www.youtube.com/watch?v=7g01DlHlQqI&ab_channel=%D0%9B%D0%B5%D0%BA%D1%82%D0%BE%D1%80%D0%B8%D0%B9%D0%A4%D0%9F%D0%9C%D0%98'

 В "го модах" использую replace, чтобы не ходить за зависимостями по сети, а подтягивался локальный пакет.

* Возможно потребуется поставить sqlite3, к примеру на Ubuntu введите две команды 

    sudo apt update
    sudo apt install sqlite3

