package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const pageTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Token</title>
	<style>
		html, body {
			height: 100%;
			margin: 0;
		}
		body {
			display: flex;
			flex-direction: column;
			align-items: stretch;
			background:
				radial-gradient(ellipse at 50% 0%, #141a2b 0%, transparent 55%),
				#0a0a0c;
			color: #fff;
			font-family: monospace;
			overflow: hidden;
		}
		/* Latar ambient: gumpalan cahaya yang bergerak pelan. */
		.ambient {
			position: fixed;
			inset: 0;
			z-index: 0;
			overflow: hidden;
			pointer-events: none;
		}
		.blob {
			position: absolute;
			width: 55vmax;
			height: 55vmax;
			border-radius: 50%;
			filter: blur(70px);
			opacity: 0.3;
			will-change: transform;
			animation: drift 28s ease-in-out infinite alternate;
		}
		.blob-1 {
			top: -20vmax;
			left: -15vmax;
			background: radial-gradient(circle, #4f46e5 0%, transparent 65%);
			animation-name: drift-a;
		}
		.blob-2 {
			top: -10vmax;
			right: -20vmax;
			background: radial-gradient(circle, #0ea5e9 0%, transparent 65%);
			animation-name: drift-b;
			animation-duration: 34s;
		}
		.blob-3 {
			bottom: -25vmax;
			left: 20vmax;
			background: radial-gradient(circle, #14b8a6 0%, transparent 65%);
			animation-name: drift-c;
			animation-duration: 40s;
		}
		@keyframes drift-a {
			from { transform: translate3d(0, 0, 0) scale(1); }
			to   { transform: translate3d(12vw, 10vh, 0) scale(1.15); }
		}
		@keyframes drift-b {
			from { transform: translate3d(0, 0, 0) scale(1.1); }
			to   { transform: translate3d(-14vw, 12vh, 0) scale(0.95); }
		}
		@keyframes drift-c {
			from { transform: translate3d(0, 0, 0) scale(0.95); }
			to   { transform: translate3d(10vw, -12vh, 0) scale(1.2); }
		}
		header, footer { position: relative; z-index: 1; text-align: center; padding: 1rem; flex: 0 0 auto; }
		header { font-size: clamp(1rem, 2.5vw, 1.5rem); }
		footer { font-size: clamp(0.8rem, 2vw, 1.1rem); color: rgba(255, 255, 255, 0.75); }
		main {
			position: relative;
			z-index: 1;
			flex: 1 1 auto;
			min-height: 0;
			display: flex;
			align-items: center;
			justify-content: center;
			padding: 0 1rem;
			overflow: hidden;
		}
		.token {
			/* Ukuran dasar; font-size sebenarnya dihitung skrip agar pas jendela. */
			font-size: 16px;
			font-weight: bold;
			text-align: center;
			word-break: break-all;
			overflow-wrap: anywhere;
			line-height: 1.1;
			max-width: 100%;
			text-shadow: 0 0 0.35em rgba(10, 10, 12, 0.85);
		}
		@media (prefers-reduced-motion: reduce) {
			.blob { animation: none; }
		}
	</style>
</head>
<body>
	<div class="ambient" aria-hidden="true">
		<span class="blob blob-1"></span>
		<span class="blob blob-2"></span>
		<span class="blob blob-3"></span>
	</div>
	<header>Token hari ini</header>
	<main><div class="token" id="token">{{.Token}}</div></main>
	<footer>Umur token: <span id="age">0</span> menit</footer>
	<script>
		const tokenView = document.getElementById("token");
		const ageView = document.getElementById("age");
		const tokenBox = document.querySelector("main");
		let tokenChangedAt = {{.ChangedAt}};

		// Sesuaikan ukuran teks token agar memenuhi jendela tanpa terpotong.
		function fitToken() {
			if (!tokenBox || !tokenView.textContent) return;
			const cs = getComputedStyle(tokenBox);
			const availW = tokenBox.clientWidth
				- parseFloat(cs.paddingLeft) - parseFloat(cs.paddingRight);
			const availH = tokenBox.clientHeight
				- parseFloat(cs.paddingTop) - parseFloat(cs.paddingBottom);
			if (availW <= 0 || availH <= 0) return;

			let lo = 4;
			let hi = Math.max(availW, availH);
			for (let i = 0; i < 20; i++) {
				const mid = (lo + hi) / 2;
				tokenView.style.fontSize = mid + "px";
				if (tokenView.scrollWidth <= availW && tokenView.scrollHeight <= availH) {
					lo = mid;
				} else {
					hi = mid;
				}
			}
			tokenView.style.fontSize = Math.floor(lo) + "px";
		}

		let fitQueued = false;
		function scheduleFit() {
			if (fitQueued) return;
			fitQueued = true;
			requestAnimationFrame(() => {
				fitQueued = false;
				fitToken();
			});
		}

		window.addEventListener("resize", scheduleFit);
		if (window.ResizeObserver) {
			new ResizeObserver(scheduleFit).observe(tokenBox);
		}
		if (document.fonts && document.fonts.ready) {
			document.fonts.ready.then(scheduleFit);
		}
		fitToken();

		setInterval(() => {
			ageView.textContent = Math.floor((Date.now() - tokenChangedAt) / 60000);
		}, 1000);
		const events = new EventSource("/events");
		events.onmessage = (event) => {
			if (tokenView.textContent !== event.data) {
				tokenView.textContent = event.data;
				tokenChangedAt = Date.now();
				ageView.textContent = "0";
				scheduleFit();
			}
		};
	</script>
</body>
</html>`

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Masukkan token: ")
	if !scanner.Scan() {
		log.Fatal("gagal membaca input")
	}
	token := strings.TrimSpace(scanner.Text())
	if token == "" {
		log.Fatal("token tidak boleh kosong")
	}
	tokenChangedAt := time.Now()

	tmpl := template.Must(template.New("page").Parse(pageTemplate))
	var tokenMu sync.RWMutex
	var clientsMu sync.Mutex
	clients := make(map[chan string]struct{})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "metode tidak diizinkan", http.StatusMethodNotAllowed)
			return
		}
		tokenMu.RLock()
		currentToken := token
		currentChangedAt := tokenChangedAt
		tokenMu.RUnlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, struct {
			Token     string
			ChangedAt int64
		}{Token: currentToken, ChangedAt: currentChangedAt.UnixMilli()}); err != nil {
			log.Println("gagal render template:", err)
		}
	})
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "metode tidak diizinkan", http.StatusMethodNotAllowed)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "stream tidak didukung", http.StatusInternalServerError)
			return
		}
		updates := make(chan string, 1)
		clientsMu.Lock()
		clients[updates] = struct{}{}
		clientsMu.Unlock()
		defer func() {
			clientsMu.Lock()
			delete(clients, updates)
			clientsMu.Unlock()
		}()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		tokenMu.RLock()
		currentToken := token
		tokenMu.RUnlock()
		fmt.Fprintf(w, "data: %s\n\n", currentToken)
		flusher.Flush()
		for {
			select {
			case <-r.Context().Done():
				return
			case updatedToken := <-updates:
				fmt.Fprintf(w, "data: %s\n\n", updatedToken)
				flusher.Flush()
			}
		}
	})

	addr := "127.0.0.1:8080"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("gagal menjalankan server: ", err)
	}
	url := "http://" + addr
	fmt.Printf("Server berjalan di %s\n", url)
	go func() {
		fmt.Print("Masukkan token baru: ")
		for scanner.Scan() {
			newToken := strings.TrimSpace(scanner.Text())
			if newToken == "" {
				fmt.Print("Token tidak boleh kosong. Masukkan token baru: ")
				continue
			}
			tokenMu.Lock()
			token = newToken
			tokenChangedAt = time.Now()
			tokenMu.Unlock()
			clientsMu.Lock()
			for client := range clients {
				select {
				case client <- newToken:
				default:
					select {
					case <-client:
					default:
					}
					client <- newToken
				}
			}
			clientsMu.Unlock()
			fmt.Print("Masukkan token baru: ")
		}
		if err := scanner.Err(); err != nil {
			log.Printf("gagal membaca token baru: %v", err)
		}
	}()
	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start(); err != nil {
		log.Printf("gagal membuka browser otomatis: %v", err)
		fmt.Printf("Buka alamat ini secara manual: %s\n", url)
	}
	log.Fatal(http.Serve(listener, nil))
}
