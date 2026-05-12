import { useEffect, useState } from 'react';
import { fetchJson } from '../api';

interface Validator {
  id: string;
  name: string;
  status: string;
  votingPower: number;
}

function Validators() {
  const [validators, setValidators] = useState<Validator[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchJson<Validator[]>('/validators')
      .then(setValidators)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load validators'));
  }, []);

  return (
    <section className="page page-data">
      <div className="card">
        <h2>Validators</h2>
        {error && <p className="error-message">{error}</p>}
        <div className="grid-list">
          {validators.length > 0 ? (
            validators.map((validator) => (
              <article key={validator.id} className="data-item">
                <strong>{validator.name}</strong>
                <span>{validator.status}</span>
                <small>Power: {validator.votingPower}</small>
              </article>
            ))
          ) : (
            <p>Loading validator data...</p>
          )}
        </div>
      </div>
    </section>
  );
}

export default Validators;
