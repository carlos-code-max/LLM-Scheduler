import React, { useState, useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import TaskList from './TaskList';
import TaskDetail from './TaskDetail';

const Tasks: React.FC = () => {
  return (
    <Routes>
      <Route path="/" element={<TaskList />} />
      <Route path="/list" element={<TaskList />} />
      <Route path="/detail/:id" element={<TaskDetail />} />
      <Route path="*" element={<Navigate to="/tasks/list" replace />} />
    </Routes>
  );
};

export default Tasks;
