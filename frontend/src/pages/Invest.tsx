import { useEffect, useState } from 'react';
import { fetchJson } from '../api';

interface Investment {
  id: string;
  name: string;
  description: string;
  returnRate: string;
}

function Invest() {
  const [items, setItems] = useState<Investment[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchJson<Investment[]>('/invest')
      .then(setItems)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load investment options'));
  }, []);

  return (
    <section className="page page-data">
      <div className="card">
        <h2>Investment Options</h2>
        {error && <p className="error-message">{error}</p>}
        <div className="grid-list">
          {items.length > 0 ? (
            items.map((item) => (
              <article key={item.id} className="data-item">
                <strong>{item.name}</strong>
                <span>{item.returnRate}</span>
                <p>{item.description}</p>
              </article>
            ))
          ) : (
            <p>Loading investment campaigns...</p>
          )}
        </div>
      </div>
    </section>
  );
}

export default Invest;
