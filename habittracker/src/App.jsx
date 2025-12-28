import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import RegisterPage from './assets/components/RegisterPage';
import LoginPage from './assets/components/LoginPage';
import Room from './assets/components/Room';
import './App.css';

function App() {
  return (
    <Router>
      <div className="App">
        <Routes>
          {/* Стартовая страница - редирект с корня на /login */}
          <Route path="/" element={<Navigate to="/login" replace />} />

          {/* Страница логина */}
          <Route path="/login" element={<LoginPage />} />

          {/* Страница регистрации */}
          <Route path="/register" element={<RegisterPage />} />

          {/* Главная страница (Room) */}
          <Route path="/room" element={<Room />} />

          {/* Fallback для несуществующих маршрутов - тоже на логин */}
          <Route path="*" element={<Navigate to="/login" replace />} />

          {/* Навигация, чтобы вернуться на главную страницу */}
          <Route path="/room/me" element={<Room />} />
          
          {/* Навигация в гости */}
          <Route path="/room/:userId" element={<Room />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;