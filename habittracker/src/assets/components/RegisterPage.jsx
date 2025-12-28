import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './RegisterPage.css';


const API_URL = 'http://localhost:8080/api';

function RegisterPage() {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
  e.preventDefault();
  
  setIsLoading(true);
  setError('');

  // Получение данных из формы
  const username = e.target.username.value;
  const password = e.target.password.value;
  const confirmPassword = e.target['confirm-password'].value;

  // Проверка паролей
  if (password !== confirmPassword) {
    setError('Пароли не совпадают!');
    setIsLoading(false);
    return;
  }

  try {
    // запрос на регистрацию
    const response = await fetch(API_URL + "/register", {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });

    const data = await response.json();
    
    if (response.ok) {
      navigate('/login');
    } else {
      alert("Похоже, ваши данные не нравятся дому!");
    }
  } catch (error) {
    alert("Возникла загадочная ошибка!");
  } finally {
    // ВАЖНО: Всегда сбрасываем состояние загрузки
    setIsLoading(false);
  }
};

  const handleLoginClick = (e) => {
    e.preventDefault();
    navigate('/login');
  };

      return (
        <div className="register-page">
          {/* Фоновое изображение */}
          <div className="background-image" />

          {/* Форма регистрации */}
          <div className="register-form-container">
            <form className="register-form" onSubmit={handleSubmit}>
              <h1 className="form-title">Создать дом</h1>

              <div className="form-group">
                <label htmlFor="username">Вас зовут...</label>
                <input
                  type="text"
                  id="username"
                  name="username"
                  placeholder="Введите имя"
                  className="pixel-input"
                  maxLength="13"
                  required
                  disabled={isLoading}
                />
              </div>

              <div className="form-group">
                <label htmlFor="password">Создайте пароль</label>
                <input
                  type="password"
                  id="password"
                  name="password"
                  placeholder="Придумайте пароль"
                  className="pixel-input"
                  minLength="5"
                  required
                  disabled={isLoading}
                />
              </div>

              <div className="form-group">
                <label htmlFor="confirm-password">Какой пароль вы создали?</label>
                <input
                  type="password"
                  id="confirm-password"
                  name="confirm-password"
                  placeholder="Повторите пароль"
                  className="pixel-input"
                  required
                  disabled={isLoading}
                />
              </div>

              {/* Картинка dwer.webp */}
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
                />
              </div>

              {/* Кнопка "зарегистрироваться" */}
              <button type="submit" 
              className="register-link pixel-button"
              disabled={isLoading}>
                ЗАРЕГИСТРИРОВАТЬСЯ
              </button>

              {/* Ссылка "У меня уже есть дом" */}
              <div className="login-link-container">
                <a href="#" onClick={handleLoginClick} className="login-link">
                  У МЕНЯ УЖЕ ЕСТЬ ДОМ
                </a>
              </div>
            </form>
          </div>
        </div>
      );
    }

export default RegisterPage;