// main.test.jsx
import React from 'react';

describe('Тестирование мэйн- точки входа', () => {
  let mockRender;
  let mockCreateRoot;
  let originalGetElementById;
  let mockGetElementById;
  let mockRootElement;

  beforeEach(() => {
    jest.clearAllMocks();
    
    // Создаем свежие моки для каждого теста
    mockRender = jest.fn();
    mockCreateRoot = jest.fn(() => ({
      render: mockRender,
    }));
    
    // Мокаем react-dom/client
    jest.doMock('react-dom/client', () => ({
      createRoot: mockCreateRoot,
    }));
    
    jest.doMock('../src/App.jsx', () => {
      const MockApp = () => <div>Mocked App</div>;
      MockApp.displayName = 'App';
      return MockApp;
    });
    
    jest.doMock('../src/index.css', () => ({}));
    
    originalGetElementById = document.getElementById;
    
    mockRootElement = {
      innerHTML: '',
    };
    mockGetElementById = jest.fn(() => mockRootElement);
    document.getElementById = mockGetElementById;
  });

  afterEach(() => {
    document.getElementById = originalGetElementById;
    jest.resetModules();
  });

  test('вызывает createRoot с правильным элементом', () => {
    require('../src/main.jsx');
    
    expect(mockGetElementById).toHaveBeenCalledWith('root');
    expect(mockCreateRoot).toHaveBeenCalledWith(mockRootElement);
  });

  test('вызывает render на созданном root', () => {
    require('../src/main.jsx');
    
    expect(mockRender).toHaveBeenCalled();
  });

  test('рендерит App внутри StrictMode', () => {
    require('../src/main.jsx');

    expect(mockRender).toHaveBeenCalled();
    
    const renderCall = mockRender.mock.calls[0];
    expect(renderCall).toBeDefined();
    
    const renderedContent = renderCall[0];
    expect(renderedContent).toBeDefined();

    expect(renderedContent.type).toBe(React.StrictMode);

    const appComponent = renderedContent.props.children;
    expect(appComponent.type.displayName).toBe('App');
  });

  test('импортирует index.css', () => {
    expect(() => require('../src/main.jsx')).not.toThrow();
  });
});


