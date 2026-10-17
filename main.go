package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

// Участники, которым можно отправлять сердечки.
var names = []string{
	"Саргсян Жасмин",
	"Адамян Марэта",
	"Айвазян Айва",
	"Манукян Анна",
	"Даниелян Жозета",
	"Тагесян Луиза",
	"Галоян Сергей",
	"Даниелян Сергей",
	"Яйлаханян Давид",
	"Арзанунц Арина",
	"Хачатрян Кристина",
	"Агаджанян Тигран",
	"Акопджанян Армен",
	"Багоян Елена",
	"Цатурян Эдгар",
	"Шахвердян Сюзи",
	"Багдасарян Акоб",
	"Абрамян Нарек",
	"Еганян Рита",
	"Сукиасян Маша",
	"Балаян Луиза",
	"Назарян Геворг",
	"Айвазян Арам",
	"Еганян Алина",
	"Нерсисян Полина",
	"Андреева Алина",
	"Хачатурян Мариам",
	"Хидирян Армине",
	"Мнеян Тамара",
	"Мелконян Диана",
	"Сейранян Анна",
	"Саарян Ева",
	"Даниелян Давид",
	"Геворгян Гоарик",
}

// Mutex не позволяет двум HTTP-запросам одновременно менять статистику.
var mu sync.Mutex
var counts = make(map[string]int)

func main() {
	// Если сервер запускался раньше, загружаем сохранённые сердечки.
	data, err := os.ReadFile("hearts.json")
	if err == nil {
		if err := json.Unmarshal(data, &counts); err != nil {
			log.Fatal("Ошибка чтения hearts.json: ", err)
		}
	} else if !os.IsNotExist(err) {
		log.Fatal("Ошибка открытия hearts.json: ", err)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/hearts", heartsHandler)
	http.HandleFunc("/api/reset", resetHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Сервер запущен: http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// GET / — возвращает HTML-страницу.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Используйте GET", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page)
}

// GET /api/hearts — получить статистику.
// POST /api/hearts — добавить одно сердечко выбранному участнику.
func heartsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		result := snapshot()
		mu.Unlock()
		sendJSON(w, result)

	case http.MethodPost:
		var request struct {
			Name string `json:"name"`
		}
		decoder := json.NewDecoder(io.LimitReader(r.Body, 1024))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "Ожидается JSON с полем name", http.StatusBadRequest)
			return
		}
		// Проверяем, что передан ровно один JSON-объект.
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			http.Error(w, "Некорректный JSON", http.StatusBadRequest)
			return
		}
		found := false
		for _, name := range names {
			if name == request.Name {
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Участник не найден", http.StatusBadRequest)
			return
		}

		mu.Lock()
		counts[request.Name]++
		data, err := json.MarshalIndent(counts, "", "  ")
		if err == nil {
			err = os.WriteFile("hearts.json", data, 0644)
		}
		if err != nil {
			counts[request.Name]--
			mu.Unlock()
			http.Error(w, "Не удалось сохранить данные", http.StatusInternalServerError)
			return
		}
		result := snapshot()
		mu.Unlock()
		sendJSON(w, result)

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Разрешены только GET и POST", http.StatusMethodNotAllowed)
	}
}

// POST /api/reset — обнулить счётчики всех участников.
// В этой учебной версии сброс доступен без авторизации — только для локальной демонстрации.
func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Разрешён только POST", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	oldCounts := counts
	counts = make(map[string]int)
	// Сохраняем пустую статистику, чтобы после перезапуска она осталась обнулённой.
	if err := os.WriteFile("hearts.json", []byte("{}\n"), 0644); err != nil {
		counts = oldCounts
		mu.Unlock()
		http.Error(w, "Не удалось сбросить статистику", http.StatusInternalServerError)
		return
	}
	result := snapshot()
	mu.Unlock()
	sendJSON(w, result)
}

type heartStats struct {
	Total  int            `json:"total"`
	Counts map[string]int `json:"counts"`
}

// Вызывать только при заблокированном mu.
func snapshot() heartStats {
	data := heartStats{Counts: make(map[string]int)}
	for _, name := range names {
		value := counts[name]
		data.Counts[name] = value
		data.Total += value
	}
	return data
}

func sendJSON(w http.ResponseWriter, value heartStats) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Println("Ошибка отправки JSON:", err)
	}
}

const page = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Сердце AK 🧡</title>
    <style>
        :root { color-scheme: light; }
        * { box-sizing: border-box; }
        body {
            margin: 0;
            min-height: 100vh;
            background: #f5c58a;
            color: #e66f1a;
            font-family: Arial, sans-serif;
        }
        .page { padding: 24px 12px 40px; }
        .intro { text-align: center; }
        .eyebrow {
            font-size: 12px;
            font-weight: 800;
            text-transform: uppercase;
            letter-spacing: .21em;
            opacity: .8;
        }
        .intro h2 { margin: 8px 0 4px; font-size: clamp(18px, 2vw, 27px); }
        .intro p { font-size: 14px; margin: 0; color: #9b4d15; }
        .scene-scroll { width: 100%; overflow-x: auto; }
        .scene {
            position: relative;
            width: min(1400px, calc(100vw - 24px));
            min-width: 1100px;
            aspect-ratio: 16 / 9;
            margin: 0 auto;
        }
        .name {
            position: absolute;
            transform: translate(-50%, -50%);
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            width: 7.4%;
            min-height: 35px;
            padding: 4px 2px;
            font-family: inherit;
            font-size: clamp(10px, 0.87vw, 13px);
            line-height: 1.12;
            font-weight: 700;
            text-align: center;
            color: #a54c11;
            background: transparent;
            border: 1px solid transparent;
            border-radius: 12px;
            cursor: pointer;
            transition: background .18s, border-color .18s, transform .18s;
            -webkit-tap-highlight-color: transparent;
        }
        .name:hover, .name:focus-visible {
            background: #fff0d8;
            border-color: #ef9a48;
            transform: translate(-50%, -50%) scale(1.1);
            outline: none;
            z-index: 3;
        }
        .name:disabled { cursor: wait; opacity: .55; }
        .count { display: block; font-size: 10px; line-height: 1.1; color: #cc6015; }
        .count:empty { display: none; }
        .center {
            position: absolute;
            left: 50%; top: 55%;
            transform: translate(-50%, -50%);
            text-align: center;
            pointer-events: none;
        }
        .heart {
            font-size: clamp(75px, 9vw, 145px);
            line-height: 1.2;
            animation: beat 1.15s ease-in-out infinite;
        }
        h1 {
            color: #e66f1a;
            font-size: clamp(22px, 2.4vw, 40px);
            margin: 4px 0 0;
            white-space: nowrap;
        }
        @keyframes beat {
            0%, 100% { transform: scale(1); }
            50% { transform: scale(1.15); }
        }
        .panel {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 18px;
            max-width: 820px;
            margin: -9px auto 0;
            padding: 14px 20px;
            border: 1px solid #ecac65;
            border-radius: 19px;
            background: #fbe0bd;
            box-shadow: 0 5px 20px #b65a1a18;
            flex-wrap: wrap;
        }
        .total { display: flex; align-items: center; gap: 10px; }
        .total strong { font-size: 28px; color: #cf5d12; }
        .total span { font-size: 13px; font-weight: bold; color: #944714; }
        .status { flex: 1; min-width: 225px; color: #944714; font-size: 13px; text-align: center; }
        .refresh {
            background: #e66f1a;
            color: white;
            border: 0;
            border-radius: 12px;
            padding: 11px 15px;
            font-weight: bold;
            font-size: 13px;
            cursor: pointer;
        }
        .refresh:hover { background: #c8540d; }
        .reset {
            background: transparent;
            color: #a54c11;
            border: 1px solid #d88a4e;
            border-radius: 12px;
            padding: 10px 14px;
            font-weight: bold;
            font-size: 13px;
            cursor: pointer;
        }
        .reset:hover { background: #ffe9d1; }
        .reset:disabled { cursor: wait; opacity: .6; }
        .fly {
            position: fixed;
            pointer-events: none;
            z-index: 20;
            font-size: 26px;
            animation: floatUp 1.1s ease-out forwards;
        }
        @keyframes floatUp {
            from { opacity: 1; transform: translate(-50%, -50%) scale(.7); }
            to { opacity: 0; transform: translate(-50%, -125px) scale(1.5); }
        }
        @media (prefers-reduced-motion: reduce) {
            *, *::before, *::after { animation: none !important; transition: none !important; }
        }
        @media (max-width: 700px) {
            .page { padding-top: 18px; }
            .intro p { padding: 0 12px; }
            .panel { margin-top: 10px; }
            .status { order: 3; flex-basis: 100%; }
        }
    </style>
</head>
<body>
    <main class="page">
        <header class="intro">
            <div class="eyebrow">Наша команда</div>
            <h2>Подари сердечко участнику</h2>
            <p>Нажми на любое имя — сердечко отправится на Go-сервер и останется в общей статистике.</p>
        </header>
        <div class="scene-scroll">
            <div class="scene">
        <button class="name" type="button" data-name="Саргсян Жасмин" style="left:50%;top:35.5%" aria-label="Подарить сердечко: Саргсян Жасмин"><span>Саргсян<br>Жасмин</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Адамян Марэта" style="left:53.6%;top:26.9%" aria-label="Подарить сердечко: Адамян Марэта"><span>Адамян<br>Марэта</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Айвазян Айва" style="left:59.1%;top:21.8%" aria-label="Подарить сердечко: Айвазян Айва"><span>Айвазян<br>Айва</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Манукян Анна" style="left:65%;top:19%" aria-label="Подарить сердечко: Манукян Анна"><span>Манукян<br>Анна</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Даниелян Жозета" style="left:71.2%;top:18.2%" aria-label="Подарить сердечко: Даниелян Жозета"><span>Даниелян<br>Жозета</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Тагесян Луиза" style="left:77.3%;top:19.1%" aria-label="Подарить сердечко: Тагесян Луиза"><span>Тагесян<br>Луиза</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Галоян Сергей" style="left:83.3%;top:21.7%" aria-label="Подарить сердечко: Галоян Сергей"><span>Галоян<br>Сергей</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Даниелян Сергей" style="left:88.8%;top:26.6%" aria-label="Подарить сердечко: Даниелян Сергей"><span>Даниелян<br>Сергей</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Яйлаханян Давид" style="left:92.7%;top:34.8%" aria-label="Подарить сердечко: Яйлаханян Давид"><span>Яйлаханян<br>Давид</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Арзанунц Арина" style="left:91.6%;top:45.2%" aria-label="Подарить сердечко: Арзанунц Арина"><span>Арзанунц<br>Арина</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Хачатрян Кристина" style="left:87.1%;top:52.7%" aria-label="Подарить сердечко: Хачатрян Кристина"><span>Хачатрян<br>Кристина</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Агаджанян Тигран" style="left:81.8%;top:58.3%" aria-label="Подарить сердечко: Агаджанян Тигран"><span>Агаджанян<br>Тигран</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Акопджанян Армен" style="left:76.2%;top:63.2%" aria-label="Подарить сердечко: Акопджанян Армен"><span>Акопджанян<br>Армен</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Багоян Елена" style="left:70.6%;top:67.7%" aria-label="Подарить сердечко: Багоян Елена"><span>Багоян<br>Елена</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Цатурян Эдгар" style="left:65%;top:72.2%" aria-label="Подарить сердечко: Цатурян Эдгар"><span>Цатурян<br>Эдгар</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Шахвердян Сюзи" style="left:59.4%;top:77%" aria-label="Подарить сердечко: Шахвердян Сюзи"><span>Шахвердян<br>Сюзи</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Багдасарян Акоб" style="left:54.1%;top:82.6%" aria-label="Подарить сердечко: Багдасарян Акоб"><span>Багдасарян<br>Акоб</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Абрамян Нарек" style="left:50%;top:90.5%" aria-label="Подарить сердечко: Абрамян Нарек"><span>Абрамян<br>Нарек</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Еганян Рита" style="left:45.9%;top:82.6%" aria-label="Подарить сердечко: Еганян Рита"><span>Еганян<br>Рита</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Сукиасян Маша" style="left:40.6%;top:77%" aria-label="Подарить сердечко: Сукиасян Маша"><span>Сукиасян<br>Маша</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Балаян Луиза" style="left:35%;top:72.2%" aria-label="Подарить сердечко: Балаян Луиза"><span>Балаян<br>Луиза</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Назарян Геворг" style="left:29.4%;top:67.7%" aria-label="Подарить сердечко: Назарян Геворг"><span>Назарян<br>Геворг</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Айвазян Арам" style="left:23.8%;top:63.2%" aria-label="Подарить сердечко: Айвазян Арам"><span>Айвазян<br>Арам</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Еганян Алина" style="left:18.2%;top:58.3%" aria-label="Подарить сердечко: Еганян Алина"><span>Еганян<br>Алина</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Нерсисян Полина" style="left:12.9%;top:52.7%" aria-label="Подарить сердечко: Нерсисян Полина"><span>Нерсисян<br>Полина</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Андреева Алина" style="left:8.5%;top:45.2%" aria-label="Подарить сердечко: Андреева Алина"><span>Андреева<br>Алина</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Хачатурян Мариам" style="left:7.3%;top:34.9%" aria-label="Подарить сердечко: Хачатурян Мариам"><span>Хачатурян<br>Мариам</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Хидирян Армине" style="left:11.2%;top:26.6%" aria-label="Подарить сердечко: Хидирян Армине"><span>Хидирян<br>Армине</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Мнеян Тамара" style="left:16.7%;top:21.7%" aria-label="Подарить сердечко: Мнеян Тамара"><span>Мнеян<br>Тамара</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Мелконян Диана" style="left:22.7%;top:19.1%" aria-label="Подарить сердечко: Мелконян Диана"><span>Мелконян<br>Диана</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Сейранян Анна" style="left:28.8%;top:18.2%" aria-label="Подарить сердечко: Сейранян Анна"><span>Сейранян<br>Анна</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Саарян Ева" style="left:35%;top:19%" aria-label="Подарить сердечко: Саарян Ева"><span>Саарян<br>Ева</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Даниелян Давид" style="left:40.9%;top:21.8%" aria-label="Подарить сердечко: Даниелян Давид"><span>Даниелян<br>Давид</span><small class="count" aria-hidden="true"></small></button>
        <button class="name" type="button" data-name="Геворгян Гоарик" style="left:46.4%;top:26.9%" aria-label="Подарить сердечко: Геворгян Гоарик"><span>Геворгян<br>Гоарик</span><small class="count" aria-hidden="true"></small></button>
                <div class="center">
                    <div class="heart">🧡</div>
                    <h1>Привет, AK!</h1>
                </div>
            </div>
        </div>
        <section class="panel" aria-label="Статистика сердечек">
            <div class="total"><strong id="total">0</strong><span>сердечек<br>отправлено</span></div>
            <div class="status" id="status" role="status" aria-live="polite">Подключаемся к серверу...</div>
            <button class="refresh" id="refresh" type="button">Обновить счётчик</button>
            <button class="reset" id="reset" type="button">Сбросить сердечки</button>
        </section>
    </main>
    <script>
        const names = document.querySelectorAll('.name');
        const total = document.getElementById('total');
        const status = document.getElementById('status');
        const refresh = document.getElementById('refresh');
        const reset = document.getElementById('reset');

        function showCounts(data) {
            total.textContent = data.total;
            for (const button of names) {
                const value = data.counts[button.dataset.name] || 0;
                button.querySelector('.count').textContent = value ? '🧡 ' + value : '';
            }
        }

        async function loadCounts(showMessage = false) {
            try {
                const response = await fetch('/api/hearts', {cache: 'no-store'});
                if (!response.ok) throw new Error('HTTP ' + response.status);
                showCounts(await response.json());
                if (showMessage) status.textContent = 'Данные обновлены. HTTP ' + response.status + ' OK';
            } catch (error) {
                status.textContent = 'Не удалось получить данные: ' + error.message;
            }
        }

        function flyHeart(button) {
            const rect = button.getBoundingClientRect();
            const particle = document.createElement('span');
            particle.className = 'fly';
            particle.textContent = '🧡';
            particle.style.left = (rect.left + rect.width / 2) + 'px';
            particle.style.top = (rect.top + rect.height / 2) + 'px';
            document.body.appendChild(particle);
            setTimeout(() => particle.remove(), 1200);
        }

        for (const button of names) {
            button.addEventListener('click', async () => {
                button.disabled = true;
                status.textContent = 'Отправляем HTTP POST-запрос...';
                try {
                    const response = await fetch('/api/hearts', {
                        method: 'POST',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify({name: button.dataset.name})
                    });
                    if (!response.ok) throw new Error('HTTP ' + response.status);
                    showCounts(await response.json());
                    flyHeart(button);
                    status.textContent = 'Сердечко для ' + button.dataset.name + ' отправлено! HTTP 200 OK';
                } catch (error) {
                    status.textContent = 'Не удалось отправить сердечко: ' + error.message;
                } finally {
                    button.disabled = false;
                }
            });
        }

        refresh.addEventListener('click', () => loadCounts(true));
        reset.addEventListener('click', async () => {
            if (!confirm('Обнулить все сердечки у всех участников?')) return;
            reset.disabled = true;
            status.textContent = 'Отправляем HTTP POST-запрос для сброса...';
            try {
                const response = await fetch('/api/reset', {method: 'POST'});
                if (!response.ok) throw new Error('HTTP ' + response.status);
                showCounts(await response.json());
                status.textContent = 'Все счётчики обнулены. HTTP 200 OK';
            } catch (error) {
                status.textContent = 'Не удалось сбросить счётчики: ' + error.message;
            } finally {
                reset.disabled = false;
            }
        });
        loadCounts();
        setInterval(() => loadCounts(), 5000);
    </script>
</body>
</html>`
