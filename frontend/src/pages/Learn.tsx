import { useEffect, useState } from 'react';
import { fetchJson } from '../api';

interface Resource {
  id: string;
  title: string;
  summary: string;
}

function Learn() {
  const [resources, setResources] = useState<Resource[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchJson<Resource[]>('/sentience/learn')
      .then(setResources)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load learning resources'));
  }, []);

  return (
    <section className="page page-data">
      <div className="card">
        <h2>Sentience Learning</h2>
        {error && <p className="error-message">{error}</p>}
        {resources.length > 0 ? (
          <div className="grid-list">
            {resources.map((item) => (
              <article key={item.id} className="data-item">
                <strong>{item.title}</strong>
                <p>{item.summary}</p>
              </article>
            ))}
          </div>
        ) : (
          <p>Loading learning resources...</p>
        )}
      </div>
    </section>
  );
}

export default Learn;
