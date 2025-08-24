import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { Layout } from 'antd';
import Sidebar from './components/Layout/Sidebar';
import Header from './components/Layout/Header';
import Dashboard from './pages/Dashboard';
import Tasks from './pages/Tasks';
import Models from './pages/Models';
import Statistics from './pages/Statistics';
import System from './pages/System';

const { Content } = Layout;

const App: React.FC = () => {
  return (
    <Layout>
      <Sidebar />
      <Layout style={{ marginLeft: 200 }}>
        <Header />
        <Content style={{ padding: '24px', minHeight: 'calc(100vh - 64px)' }}>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/tasks/*" element={<Tasks />} />
            <Route path="/models/*" element={<Models />} />
            <Route path="/statistics/*" element={<Statistics />} />
            <Route path="/system/*" element={<System />} />
          </Routes>
        </Content>
      </Layout>
    </Layout>
  );
};

export default App;
