import { useNavigate } from 'react-router-dom';

function Home() {
  const navigate = useNavigate();

  return (
    <section className="page page-home">
      <div className="hero-card">
        <h1>Sphere Online</h1>
        <p>Explore blockchain data, validator status, learning resources, investment products, and Minecraft community services.</p>
        <div className="hero-actions">
          <button onClick={() => navigate('/dashboard')}>View Dashboard</button>
          <button onClick={() => navigate('/auth')} className="secondary">
            Sign in / Register
          </button>
        </div>
      </div>
    </section>
  );
}

export default Home;
