import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import ServiceListPage from './pages/ServiceListPage';
import ServiceCreateWizard from './pages/ServiceCreateWizard';
import ServiceDetailPage from './pages/ServiceDetailPage';
import ReleasePage from './pages/ReleasePage';
import ObservabilityPage from './pages/ObservabilityPage';

const App: React.FC = () => (
  <Routes>
    <Route path="/" element={<Layout />}>
      <Route index element={<Navigate to="/services" replace />} />
      <Route path="services" element={<ServiceListPage />} />
      <Route path="services/create" element={<ServiceCreateWizard />} />
      <Route path="services/:id" element={<ServiceDetailPage />} />
      <Route path="releases" element={<ReleasePage />} />
      <Route path="observability" element={<ObservabilityPage />} />
    </Route>
  </Routes>
);

export default App;
