import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import LoginPage from '../src/assets/components/LoginPage'; 

import '@testing-library/jest-dom';

global.fetch = jest.fn();

// Мок для useNavigate
const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
}));

// Мок для alert
global.alert = jest.fn();

describe('LoginPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
    mockNavigate.mockClear();
    global.alert.mockClear();
  });

 
  test('количество буков не осилили', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    // Вводим слишком длинное имя
    const longUsername = 'a'.repeat(14);
    fireEvent.change(screen.getByLabelText(/имя/i), {
      target: { name: 'username', value: longUsername }
    });
    
    fireEvent.change(screen.getByLabelText(/пароль/i), {
      target: { name: 'password', value: 'password' }
    });

    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    // Проверяем ошибку
    await waitFor(() => {
      expect(screen.getByText(/в имени много букв\. не осилили\.\.\./i)).toBeInTheDocument();
    });
    
    expect(fetch).not.toHaveBeenCalled();
  });

  test('показывает ошибку сети при сбое запроса', async () => {
    fetch.mockRejectedValueOnce(new Error('Network error'));

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText(/имя/i), {
      target: { name: 'username', value: 'testuser' }
    });
    
    fireEvent.change(screen.getByLabelText(/пароль/i), {
      target: { name: 'password', value: 'password' }
    });

    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(screen.getByText(/возникла загадочная ошибка!/i)).toBeInTheDocument();
    });
  });

  test('очищает ошибку при изменении поля ввода', async () => {

    fetch.mockResolvedValueOnce({
      ok: false,
      json: async () => ({ message: 'Invalid credentials' })
    });

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText(/имя/i), {
      target: { name: 'username', value: 'wronguser' }
    });
    
    fireEvent.change(screen.getByLabelText(/пароль/i), {
      target: { name: 'password', value: 'wrongpass' }
    });

    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(screen.getByText(/мы таких людей не знаем/i)).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText(/имя/i), {
      target: { name: 'username', value: 'newuser' }
    });

    await waitFor(() => {
      expect(screen.queryByText(/мы таких людей не знаем/i)).not.toBeInTheDocument();
    });
  });

  test('отображает изображение с правильными атрибутами', () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    const image = screen.getByAltText(/декоративное изображение/i);
    
    expect(image).toBeInTheDocument();
    expect(image).toHaveAttribute('src', './public/Images/dwer.webp');
    expect(image).toHaveStyle({
      width: '150px',
      height: '150px',
      objectFit: 'contain'
    });
  });

  test('кнопка логина заблокирована во время загрузки', async () => {
    fetch.mockImplementationOnce(() => 
      new Promise(resolve => 
        setTimeout(() => resolve({
          ok: true,
          json: async () => ({ token: 'token' })
        }), 100)
      )
    );

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText(/имя/i), {
      target: { name: 'username', value: 'test' }
    });
    
    fireEvent.change(screen.getByLabelText(/пароль/i), {
      target: { name: 'password', value: 'test' }
    });

    const submitButton = screen.getByRole('button', { name: /войти/i });
    fireEvent.click(submitButton);

    // Проверяем, что кнопка заблокирована
    await waitFor(() => {
      expect(submitButton).toBeDisabled();
      expect(submitButton).toHaveTextContent(/загрузка/i);
    });
  });

  test('поля ввода заблокированы во время загрузки', async () => {
    fetch.mockImplementationOnce(() => 
      new Promise(resolve => 
        setTimeout(() => resolve({
          ok: true,
          json: async () => ({ token: 'token' })
        }), 100)
      )
    );

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    const usernameInput = screen.getByLabelText(/имя/i);
    const passwordInput = screen.getByLabelText(/пароль/i);
    const submitButton = screen.getByRole('button', { name: /войти/i });

    fireEvent.change(usernameInput, {
      target: { name: 'username', value: 'test' }
    });
    
    fireEvent.change(passwordInput, {
      target: { name: 'password', value: 'test' }
    });

    fireEvent.click(submitButton);
    await waitFor(() => {
      expect(usernameInput).toBeDisabled();
      expect(passwordInput).toBeDisabled();
    });
  });


  test('не отправляет форму при пустых полях', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    const submitButton = screen.getByRole('button', { name: /войти/i });

    fireEvent.click(submitButton);

    expect(fetch).not.toHaveBeenCalled();
  });

  test('успешный логин сохраняет токен и перенаправляет', async () => {
  const mockToken = 'test-jwt-token';
  fetch.mockResolvedValueOnce({
    ok: true,
    json: async () => ({ token: mockToken })
  });

  render(
    <MemoryRouter>
      <LoginPage />
    </MemoryRouter>
  );

  fireEvent.change(screen.getByLabelText(/имя/i), {
    target: { name: 'username', value: 'testuser' }
  });
  
  fireEvent.change(screen.getByLabelText(/пароль/i), {
    target: { name: 'password', value: 'correctpass' }
  });

  fireEvent.click(screen.getByRole('button', { name: /войти/i }));

  await waitFor(() => {
    expect(localStorage.setItem).toHaveBeenCalledWith('token', mockToken);
    expect(mockNavigate).toHaveBeenCalledWith('/room');
  });
});
 
 
});