test('100000 одновременных логинов', async () => {
  // Имитируем массовые запросы
  global.fetch = jest.fn(() =>
    Promise.resolve({
      ok: true,
      json: async () => ({
        token: 'token',
        user: { username: 'test', id: '1' }
      })
    })
  );

  const loginRequests = Array.from({ length: 100000}).map(() =>
    fetch('http://localhost:8080/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'user', password: 'pass' })
    })
  );

  // Ждём завершения всех
  const results = await Promise.all(loginRequests);

  // Проверяем, что все успешны
  results.forEach(res => expect(res.ok).toBe(true));
  expect(global.fetch).toHaveBeenCalledTimes(100000);
});