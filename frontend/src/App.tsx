import { lazy, Suspense } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';

import { useAuth } from './auth';
import Layout from './components/Layout';
import { Spinner } from './components/ui';
import Contracts from './pages/Contracts';
import Dashboard from './pages/Dashboard';
import Documents from './pages/Documents';
import Login from './pages/Login';
import Meters from './pages/Meters';
import Payments from './pages/Payments';
import Properties from './pages/Properties';
import PropertyDetails from './pages/PropertyDetails';
import Tenants from './pages/Tenants';
const LandingPage = lazy(() => import('./landing/LandingPage'));

function LandingRoute() {
  return <Suspense fallback={<div className="centered"><Spinner /></div>}><LandingPage /></Suspense>;
}

export default function App() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="centered">
        <Spinner />
      </div>
    );
  }

  if (!user) {
    return (
      <Routes>
        <Route path="/landing/*" element={<LandingRoute />} />
        <Route path="/login" element={<Login />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    );
  }

  return (
    <Routes>
      <Route path="/landing/*" element={<LandingRoute />} />
      <Route path="/login" element={<Navigate to="/" replace />} />
      <Route element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="properties" element={<Properties />} />
        <Route path="properties/:id" element={<PropertyDetails />} />
        <Route path="tenants" element={<Tenants />} />
        <Route path="contracts" element={<Contracts />} />
        <Route path="payments" element={<Payments />} />
        <Route path="meters" element={<Meters />} />
        <Route path="documents" element={<Documents />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}
