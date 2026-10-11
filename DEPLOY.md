# Deploying Faultline Live — Render (backend) + Vercel (frontend)

This guide takes the app from this folder to two public URLs with **zero code
changes**. Read it top to bottom; every step is required and in order.

Architecture after deploy:

```
Vercel (frontend, Vite SPA)  ──HTTPS──▶  Render (backend, Go API)  ──▶ Neon Postgres
                                              │──▶ Gemini API (investigations)
                                              └──▶ offline rules fallback (always works)
```

> Security note: the hosted demo runs with `ALLOW_PUBLIC_DEMO=true`, so anyone
> with the URL can drive simulations. The sims stay bounded (10 min expiry,
> 50% error cap). Use a throwaway database and rotate `JWT_SECRET` /
> `GEMINI_API_KEY` afterwards if the URLs were shared publicly.

---

## 0. Prerequisites (5 minutes)

1. A GitHub account.
2. A Render account (sign up with GitHub): https://render.com
3. A Vercel account (sign up with GitHub): https://vercel.com
4. Your Neon Postgres connection string (the `DATABASE_URL` already in
   `backend/.env` works — reuse it).
5. Your Google Gemini API key (already saved in local `backend/.env` as
   `GEMINI_API_KEY`). Get/verify keys at https://aistudio.google.com/apikey

## 1. Push this folder to GitHub

The deploy root is the **`faultline-live-demo/`** folder (it contains
`render.yaml`, `backend/` and `frontend/`).

```powershell
cd faultline-live-demo
git init
git add .
git commit -m "Deploy Faultline Live to Render + Vercel"
git branch -M main
git remote add origin https://github.com/<your-username>/faultline-live.git
git push -u origin main
```

> `.env` files are gitignored and will NOT be pushed. Secrets are entered in
> the Render/Vercel dashboards instead.

## 2. Deploy the backend on Render (Blueprint)

1. Render Dashboard → **New +** → **Blueprint** → select the repo you just
   pushed. Render reads `render.yaml` and shows one service:
   `faultline-live-api`.
2. Before clicking **Apply**, fill the secrets Render asks for:
   | Key | Value |
   |-----|-------|
   | `DATABASE_URL` | your Neon string, e.g. `postgresql://…@….neon.tech/neondb?sslmode=require` |
   | `GEMINI_API_KEY` | your Gemini API key |
   | `FRONTEND_URL` | leave blank for now — fill after step 3 (your `https://….vercel.app`) |
   | `CORS_ALLOWED_ORIGINS` | leave blank for now — same Vercel URL later |
   | `JWT_SECRET` | auto-generated, keep it |
3. Click **Apply**. Render builds (`go build -o server .`) and starts
   (`./server`). `AUTO_MIGRATE=true` creates all tables automatically —
   no manual migration step.
4. Wait for **Live**. Note your backend URL, e.g.
   `https://faultline-live-api.onrender.com`.
5. Verify in the browser:
   - `https://<your-api>.onrender.com/api/health` → `{"data":{"status":"ok",…}}`
   - `https://<your-api>.onrender.com/api/ready` → checks include
     `"gemini":{"status":"ok",…}` (and no `ollama` entry — correct, there is
     no Ollama daemon on Render).

   Troubleshooting:
   - `database connectivity check failed` → the `DATABASE_URL` is wrong or
     Neon doesn't allow the connection. Confirm `sslmode=require`.
   - `Gemini API rejected the request` → wrong key or wrong `GEMINI_MODEL`.
     Fix the env var, **Save**, Render redeploys automatically.
   - Build fails on Go version → set env `GOTOOLCHAIN=local`? No — Render
     uses the toolchain in `go.mod` (1.26.1). A build error log names the
     real cause (usually a missing env var, not Go itself).

## 3. Deploy the frontend on Vercel

1. Vercel Dashboard → **Add New → Project** → import the same repo.
2. When asked for the project root, set **Root Directory = `frontend`**
   (this is the most common mistake — the repo root will fail to build).
3. Framework preset: **Vite**. Build command `npm run build`, output
   directory `dist` (both already in `frontend/vercel.json`).
4. Open **Environment Variables** and add exactly one variable:
   | Key | Value |
   |-----|-------|
   | `VITE_API_BASE_URL` | your Render URL from step 2, e.g. `https://faultline-live-api.onrender.com` (no trailing slash) |
5. Click **Deploy**. Note your frontend URL, e.g.
   `https://faultline-live.vercel.app`.

## 4. Connect frontend ↔ backend (CORS)

Back in **Render → faultline-live-api → Environment**, set both:

```
FRONTEND_URL=https://<your-app>.vercel.app
CORS_ALLOWED_ORIGINS=https://<your-app>.vercel.app
```

Click **Save Changes** — Render redeploys (~2 min). This is the step that
fixes the classic blank-dashboard / `Failed to fetch` / CORS error.

## 5. End-to-end verification (do all four)

1. Open `https://<your-app>.vercel.app/ops` → system banner loads, telemetry
   appears after a few seconds.
2. **Simulations**: inject latency → Generate traffic → an incident fires.
3. **Investigate**: open the incident → Start Investigation → you get a
   Gemini-written analysis (model badge `gemini:gemini-2.0-flash`). If Gemini
   is unreachable you still get offline analysis — never a bare error.
4. **Defend**: on a `traffic_spike` incident enable the rate limiter, flood
   again, confirm excess calls return HTTP 429.

## 6. Env-var cheat sheet

Backend (Render):
`SERVER_HOST=0.0.0.0` · `PORT` (injected by Render) · `DATABASE_URL` ·
`JWT_SECRET` · `FAULTLINE_DEMO_MODE=true` · `ALLOW_PUBLIC_DEMO=true` ·
`OPS_AUTH_REQUIRED=false` · `FAULTS_ENABLED=true` · `OLLAMA_ENABLED=false` ·
`GEMINI_ENABLED=true` · `GEMINI_API_KEY` · `GEMINI_MODEL=gemini-2.0-flash` ·
`AUTO_MIGRATE=true` · `FRONTEND_URL` + `CORS_ALLOWED_ORIGINS` (Vercel URL).

Frontend (Vercel): `VITE_API_BASE_URL` (Render URL). Nothing else.

Local dev is unchanged: `backend/.env` (Ollama on, Gemini key already
stored) + `go run .`, and `frontend` with empty `VITE_API_BASE_URL` (Vite
proxy). See `backend/.env.example` and `frontend/.env.example`.
