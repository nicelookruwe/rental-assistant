import { useState } from 'react';
import type { FormEvent } from 'react';

import { useAuth } from '../auth';
import { Button, ErrorNote, Field, Input } from '../components/ui';

export default function Login() {
  const { signIn, signUp } = useAuth();
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState<unknown>(null);
  const [busy, setBusy] = useState(false);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setError(null);

    if (mode === 'register' && password !== confirm) {
      setError(new Error('Пароли не совпадают'));
      return;
    }
    if (password.length < 8) {
      setError(new Error('Пароль должен содержать не менее 8 символов'));
      return;
    }

    setBusy(true);
    try {
      if (mode === 'login') {
        await signIn(email.trim(), password);
      } else {
        await signUp(email.trim(), password);
      }
    } catch (err) {
      setError(err);
    } finally {
      setBusy(false);
    }
  }

  function switchMode() {
    setMode((m) => (m === 'login' ? 'register' : 'login'));
    setError(null);
    setPassword('');
    setConfirm('');
  }

  return (
    <div className="auth">
      <aside className="auth-aside">
        <div>
          <span className="brand">
            <span className="brand-mark">RA</span>
            <span className="brand-text">
              Rental <b>Assistant</b>
            </span>
          </span>
          <h2>Порядок в аренде начинается с одного места</h2>
          <p>Квартиры, арендаторы, договоры, платежи и показания счётчиков — в одном кабинете.</p>
          <ul className="auth-features">
            <li>Контроль сроков договоров и напоминания</li>
            <li>История арендных и коммунальных платежей</li>
            <li>Показания счётчиков без потери истории</li>
            <li>Документы и фото по каждой квартире</li>
          </ul>
        </div>
        <span className="muted" style={{ color: '#94a3b8' }}>
          MVP · 1–3 квартиры
        </span>
      </aside>

      <main className="auth-main">
        <form className="auth-card" onSubmit={submit}>
          <div>
            <h1>{mode === 'login' ? 'Вход в кабинет' : 'Создание кабинета'}</h1>
            <p className="page-subtitle">
              {mode === 'login'
                ? 'Введите email и пароль, чтобы продолжить'
                : 'Регистрация занимает минуту — данные видны только вам'}
            </p>
          </div>

          <Field label="Email">
            <Input
              type="email"
              autoComplete="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
            />
          </Field>

          <Field label="Пароль" hint={mode === 'register' ? 'Не короче 8 символов' : undefined}>
            <Input
              type="password"
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
            />
          </Field>

          {mode === 'register' ? (
            <Field label="Пароль ещё раз">
              <Input
                type="password"
                autoComplete="new-password"
                required
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                placeholder="••••••••"
              />
            </Field>
          ) : null}

          <ErrorNote error={error} />

          <Button type="submit" block loading={busy}>
            {mode === 'login' ? 'Войти' : 'Зарегистрироваться'}
          </Button>

          <p className="auth-switch">
            {mode === 'login' ? 'Ещё нет кабинета? ' : 'Уже зарегистрированы? '}
            <button type="button" onClick={switchMode}>
              {mode === 'login' ? 'Создать' : 'Войти'}
            </button>
          </p>
        </form>
      </main>
    </div>
  );
}
