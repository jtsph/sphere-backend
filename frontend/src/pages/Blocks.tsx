import { useEffect, useState } from 'react';
import { fetchJson } from '../api';

interface Block {
  id: string;
  height: number;
  timestamp: string;
  validator: string;
}

function Blocks() {
  const [blocks, setBlocks] = useState<Block[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchJson<Block[]>('/blocks')
      .then(setBlocks)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load blocks'));
  }, []);

  return (
    <section className="page page-data">
      <div className="card">
        <h2>Recent Blocks</h2>
        {error && <p className="error-message">{error}</p>}
        <div className="grid-list">
          {blocks.length > 0 ? (
            blocks.map((block) => (
              <article key={block.id} className="data-item">
                <strong>#{block.height}</strong>
                <span>{block.validator}</span>
                <small>{new Date(block.timestamp).toLocaleString()}</small>
              </article>
            ))
          ) : (
            <p>Loading block data...</p>
          )}
        </div>
      </div>
    </section>
  );
}

export default Blocks;
