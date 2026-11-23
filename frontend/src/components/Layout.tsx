import { useEffect, useState } from 'react';
import { NavLink, Outlet, useLocation } from 'react-router-dom';

import { useAuth } from '../auth';

const navItems = [
  { to: '/', label: 'Обзор', icon: 'grid', end: true },
  { to: '/properties', label: 'Квартиры', icon: 'home' },
  { to: '/tenants', label: 'Арендаторы', icon: 'users' },
  { to: '/contracts', label: 'Договоры', icon: 'doc' },
  { to: '/payments', label: 'Платежи', icon: 'wallet' },
  { to: '/meters', label: 'Счётчики', icon: 'gauge' },
  { to: '/documents', label: 'Документы', icon: 'folder' },
];

const paths: Record<string, string> = {
  grid: 'M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z',
  home: 'M3 10.5 12 3l9 7.5M5 9.5V21h14V9.5',
  users: 'M16 20v-1.5a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4V20M9 10.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7zM22 20v-1.5a4 4 0 0 0-3-3.87M16 3.6a4 4 0 0 1 0 7.3',
  doc: 'M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8zM14 3v5h5M9 13h6M9 17h6',
  wallet: 'M3 8a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2zM3 10h18M16 14h2',
  gauge: 'M12 21a9 9 0 1 1 9-9M12 12l4.5-4.5',
  folder: 'M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z',
};

function NavIcon({ name }: { name: string }) {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
      <path d={paths[name] ?? paths.doc} fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export default function Layout() {
  const { user, signOut } = useAuth();
  const location = useLocation();
  const [open, setOpen] = useState(false);

  useEffect(() => setOpen(false), [location.pathname]);

  return (
    <div className="shell">
      <aside className={`sidebar${open ? ' sidebar-open' : ''}`}>
        <div className="brand">
          <span className="brand-mark">RA</span>
          <span className="brand-text">
            Rental <b>Assistant</b>
          </span>
        </div>

        <nav className="nav">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) => `nav-link${isActive ? ' active' : ''}`}
            >
              <NavIcon name={item.icon} />
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>

        <div className="sidebar-foot">
          <span className="user-email" title={user?.email}>
            {user?.email}
          </span>
          <button type="button" className="btn btn-subtle btn-block" onClick={signOut}>
            Выйти
          </button>
        </div>
      </aside>

      {open ? <div className="sidebar-scrim" onClick={() => setOpen(false)} /> : null}

      <div className="content">
        <header className="mobile-bar">
          <button type="button" className="icon-btn" onClick={() => setOpen((v) => !v)} aria-label="Меню">
            ☰
          </button>
          <span className="mobile-title">Rental Assistant</span>
        </header>
        <main className="page">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
