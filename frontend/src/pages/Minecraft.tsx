import { useEffect, useState } from 'react';
import { fetchJson } from '../api';

interface MinecraftStatus {
  serverName: string;
  status: string;
  onlinePlayers: number;
}

function Minecraft() {
  const [status, setStatus] = useState<MinecraftStatus | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchJson<MinecraftStatus>('/minecraft')
      .then(setStatus)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load Minecraft status'));
  }, []);

  return (
    <section className="page page-data">
      <div className="card">
        <h2>Minecraft Server Status</h2>
        {error && <p className="error-message">{error}</p>}
        {status ? (
          <div className="status-box">
            <p><strong>Server:</strong> {status.serverName}</p>
            <p><strong>Status:</strong> {status.status}</p>
            <p><strong>Online Players:</strong> {status.onlinePlayers}</p>
          </div>
        ) : (
          <p>Loading server status...</p>
        )}
      </div>
    </section>
  );
}

export default Minecraft;
