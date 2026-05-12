import { useEffect, useState } from 'react';
import { fetchJson } from '../api';

interface DashboardData {
  message: string;
}

function Dashboard() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchJson<DashboardData>('/dashboard')
      .then(setData)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load dashboard'));
  }, []);

  return (
    <section className="page page-dashboard">
      <div className="card">
        <h2>Dashboard</h2>
        {error && <p className="error-message">{error}</p>}
        {data ? <p>{data.message}</p> : <p>Loading dashboard data...</p>}
      </div>
    </section>
  );
}

export default Dashboard;
