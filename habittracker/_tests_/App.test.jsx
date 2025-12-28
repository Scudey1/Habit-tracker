// App.test.jsx
import React from 'react';
import { render, screen } from '@testing-library/react';
import App from '../src/App';

// Мокаем react-router-dom
jest.mock('react-router-dom', () => {
  const actual = jest.requireActual('react-router-dom');
  return {
    ...actual,
    BrowserRouter: ({ children }) => <div data-testid="browser-router">{children}</div>,
    Routes: ({ children }) => <div data-testid="routes">{children}</div>,
    Route: ({ element }) => <div data-testid="route">{element}</div>,
    Navigate: ({ to }) => <div data-testid={`navigate-to-${to}`}>Navigate to {to}</div>,
  };
});

// Мокаем дочерние компоненты
jest.mock('../src/assets/components/RegisterPage', () => () => (
  <div data-testid="register-page">Register Page</div>
));

jest.mock('../src/assets/components/LoginPage', () => () => (
  <div data-testid="login-page">Login Page</div>
));

jest.mock('../src/assets/components/Room', () => () => (
  <div data-testid="room-page">Room Page</div>
));

describe('App компонентики', () => {
  test('рендерит компоненты и роутер', () => {
    render(<App />);
    
    expect(screen.getByTestId('browser-router')).toBeInTheDocument();
    expect(screen.getByTestId('routes')).toBeInTheDocument();
  });

  test('содержит все маршруты', () => {
    const { container } = render(<App />);
    
    // Проверяем структуру
    expect(container.querySelector('.App')).toBeInTheDocument();
  });
});