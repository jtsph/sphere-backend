import { FormEvent, useState } from 'react';
import { fetchJson } from '../api';

function Auth() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [status, setStatus] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    try {
      const path = mode === 'login' ? '/auth/login' : '/auth/register';
      const result = await fetchJson<{ message: string }>(path, {
        method: 'POST',
        body: JSON.stringify({ email, password })
      });
      setStatus(result.message || 'Success');
    } catch (error) {
      setStatus(error instanceof Error ? error.message : 'Request failed');
    }
  }

  return (
    <section className="page page-auth">
      <div className="card auth-card">
        <h2>{mode === 'login' ? 'Sign In' : 'Register'}</h2>
        <p>Use your Sphere account or create one to access the dashboard endpoints.</p>
        <form onSubmit={handleSubmit} className="form-grid">
          <label>
            Email
            <input type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
          </label>
          <label>
            Password
            <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} required />
          </label>
          <div className="form-actions">
            <button type="submit">{mode === 'login' ? 'Log In' : 'Create Account'}</button>
            <button type="button" className="secondary" onClick={() => setMode(mode === 'login' ? 'register' : 'login')}>
              {mode === 'login' ? 'Switch to Register' : 'Switch to Login'}
            </button>
          </div>
        </form>
        {status ? <div className="status-message">{status}</div> : null}
      </div>
    </section>
  );
}

export default Auth;
