import { NavLink, Route, Routes } from 'react-router-dom';
import Home from './pages/Home';
import Auth from './pages/Auth';
import Dashboard from './pages/Dashboard';
import Blocks from './pages/Blocks';
import Validators from './pages/Validators';
import Learn from './pages/Learn';
import Invest from './pages/Invest';
import Minecraft from './pages/Minecraft';

const navItems = [
  { label: 'Home', path: '/' },
  { label: 'Dashboard', path: '/dashboard' },
  { label: 'Blocks', path: '/blocks' },
  { label: 'Validators', path: '/validators' },
  { label: 'Learn', path: '/learn' },
  { label: 'Invest', path: '/invest' },
  { label: 'Minecraft', path: '/minecraft' },
  { label: 'Auth', path: '/auth' }
];

function App() {
  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand">Sphere Online</div>
        <nav className="nav-links">
          {navItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </header>

      <main className="content">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/auth" element={<Auth />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/blocks" element={<Blocks />} />
          <Route path="/validators" element={<Validators />} />
          <Route path="/learn" element={<Learn />} />
          <Route path="/invest" element={<Invest />} />
          <Route path="/minecraft" element={<Minecraft />} />
        </Routes>
      </main>

      <footer className="footer">Sphere Online frontend • powered by React + Vite</footer>
    </div>
  );
}

export default App;
