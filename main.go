package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Hello, AK</title>
    <style>
        body {
            background: #f5c58a;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            font-family: Arial, sans-serif;
        }

        .scene {
            position: relative;
            width: min(1600px, 98vw, 170vh);
            aspect-ratio: 16 / 9;
        }

        .name {
            position: absolute;
            transform: translate(-50%, -50%);
            width: 6%;
            font-size: clamp(8px, 0.75vw, 12px);
            line-height: 1.25;
            font-weight: bold;
            text-align: center;
            color: #e66f1a;
            transition: transform 0.2s;
        }

        .name:hover {
            transform: translate(-50%, -50%) scale(1.3);
            z-index: 2;
        }

        .container {
            position: absolute;
            left: 50%;
            top: 55%;
            transform: translate(-50%, -50%);
            text-align: center;
        }

        .heart {
            font-size: clamp(70px, 9vw, 150px);
            line-height: 1.2;
            animation: beat 1s infinite;
        }

        h1 {
            color: #e66f1a;
            font-size: clamp(22px, 2.4vw, 40px);
            margin: 10px 0 0;
            white-space: nowrap;
        }

        @keyframes beat {
            0%, 100% { transform: scale(1); }
            50% { transform: scale(1.2); }
        }
    </style>
</head>
<body>
    <div class="scene">
        <span class="name" style="left:50%;top:35.5%">Саргсян<br>Жасмин</span>
        <span class="name" style="left:53.6%;top:26.9%">Адамян<br>Марэта</span>
        <span class="name" style="left:59.1%;top:21.8%">Айвазян<br>Айва</span>
        <span class="name" style="left:65%;top:19%">Манукян<br>Анна</span>
        <span class="name" style="left:71.2%;top:18.2%">Даниелян<br>Жозета</span>
        <span class="name" style="left:77.3%;top:19.1%">Тагесян<br>Луиза</span>
        <span class="name" style="left:83.3%;top:21.7%">Галоян<br>Сергей</span>
        <span class="name" style="left:88.8%;top:26.6%">Даниелян<br>Сергей</span>
        <span class="name" style="left:92.7%;top:34.8%">Яйлаханян<br>Давид</span>
        <span class="name" style="left:91.6%;top:45.2%">Арзанунц<br>Арина</span>
        <span class="name" style="left:87.1%;top:52.7%">Хачатрян<br>Кристина</span>
        <span class="name" style="left:81.8%;top:58.3%">Агаджанян<br>Тигран</span>
        <span class="name" style="left:76.2%;top:63.2%">Акопджанян<br>Армен</span>
        <span class="name" style="left:70.6%;top:67.7%">Багоян<br>Елена</span>
        <span class="name" style="left:65%;top:72.2%">Цатурян<br>Эдгар</span>
        <span class="name" style="left:59.4%;top:77%">Шахвердян<br>Сюзи</span>
        <span class="name" style="left:54.1%;top:82.6%">Багдасарян<br>Акоб</span>
        <span class="name" style="left:50%;top:90.5%">Абрамян<br>Нарек</span>
        <span class="name" style="left:45.9%;top:82.6%">Еганян<br>Рита</span>
        <span class="name" style="left:40.6%;top:77%">Сукиасян<br>Маша</span>
        <span class="name" style="left:35%;top:72.2%">Балаян<br>Луиза</span>
        <span class="name" style="left:29.4%;top:67.7%">Назарян<br>Геворг</span>
        <span class="name" style="left:23.8%;top:63.2%">Айвазян<br>Арам</span>
        <span class="name" style="left:18.2%;top:58.3%">Еганян<br>Алина</span>
        <span class="name" style="left:12.9%;top:52.7%">Нерсисян<br>Полина</span>
        <span class="name" style="left:8.5%;top:45.2%">Андреева<br>Алина</span>
        <span class="name" style="left:7.3%;top:34.9%">Хачатурян<br>Мариам</span>
        <span class="name" style="left:11.2%;top:26.6%">Хидирян<br>Армине</span>
        <span class="name" style="left:16.7%;top:21.7%">Мнеян<br>Тамара</span>
        <span class="name" style="left:22.7%;top:19.1%">Мелконян<br>Диана</span>
        <span class="name" style="left:28.8%;top:18.2%">Сейранян<br>Анна</span>
        <span class="name" style="left:35%;top:19%">Саарян<br>Ева</span>
        <span class="name" style="left:40.9%;top:21.8%">Даниелян<br>Давид</span>
        <span class="name" style="left:46.4%;top:26.9%">Геворгян<br>Гоарик</span>

        <div class="container">
            <div class="heart">🧡</div>
            <h1>Привет, AK!</h1>
        </div>
    </div>
</body>
</html>`)
	})

	fmt.Println("Сервер запущен: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
