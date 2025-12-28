import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './RegisterPage.css';

const API_URL = 'http://localhost:8080/api';

function LoginPage() {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [formData, setFormData] = useState({
    username: '',
    password: ''
  });

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    // Очищаем ошибку при изменении поля
    if (error) setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    const { username, password } = formData;

    if (username.length > 13) {
      setError('В имени много букв. Не осилили...');
      setIsLoading(false);
      return;
    }

    try {
      const response = await fetch(API_URL + "/login", {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      const data = await response.json();
      if (response.ok && data.token) {
        localStorage.setItem('token', data.token);
        navigate('/room');
      } else {
        setError("Мы таких людей не знаем!");
      }
    } catch (error) {
      setError("Возникла загадочная ошибка!");
    } finally {
      setIsLoading(false);
    }
  };

  const handleRegisterClick = (e) => {
    e.preventDefault();
    navigate('/register');
  };

  return (
    <div className="register-page">
      <div className="background-image" />

      <div className="register-form-container">
        <form className="register-form" onSubmit={handleSubmit}>
          <h1 className="form-title">ВОЙТИ В ДОМ</h1>

          {/* Отображение ошибки */}
          {error && (
            <div className="error-message" style={{
              color: 'red',
              backgroundColor: '#ffe6e6',
              padding: '10px',
              margin: '10px 0',
              borderRadius: '5px',
              border: '1px solid red',
              textAlign: 'center',
              
            }}>
              {error}
            </div>
          )}

          <div className="form-group">
            <label htmlFor="username">ИМЯ</label>
            <input
              type="text"
              id="username"
              name="username"
              value={formData.username}
              onChange={handleInputChange}
              maxLength="13"
              placeholder="Введите имя"
              className="pixel-input"
              required
              disabled={isLoading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">ПАРОЛЬ</label>
            <input
              type="password"
              id="password"
              name="password"
              value={formData.password}
              onChange={handleInputChange}
              placeholder="Введите пароль"
              className="pixel-input"
              required
              disabled={isLoading}
            />
          </div>

          <div style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            margin: '10px 0'
          }}>
            <img
              style={{
                width: '150px',
                height: '150px',
                objectFit: 'contain',
                backgroundColor: '#f0f0f0',
                padding: '2px',
                display: 'block'
              }}
              src="./public/Images/dwer.webp"
              alt="декоративное изображение"
            />
          </div>

          <button 
            type="submit"
            className="register-link pixel-button"
            disabled={isLoading}
          >
            {isLoading ? 'Загрузка...' : 'Войти'}
          </button>

          <div className="login-link-container">
            <a href="#" onClick={handleRegisterClick} className="login-link">
              У МЕНЯ ЕЩЕ НЕТ ДОМА
            </a>
          </div>
        </form>
      </div>
    </div>
  );
}

export default LoginPage;