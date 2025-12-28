import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import './Room.css';

const API_URL = 'http://localhost:8080/api';

{/*
  Позиция каждой мебели в комнате
*/}
const furnituree_pos = {
  Television: { left: '66%', top: '15%', scale: '0.25' },
  Armchair: { left: '-15%', top: '0%', scale: '0.20' },
  Bed: { left: '7%', top: '2%', scale: '0.2' },
  Bookshelf: { left: '-3.5%', top: '12%', scale: '0.2' },
  Chair: { left: '37%', top: '32%', scale: '0.20' },
  Desk: { left: '-9%', top: '30%', scale: '0.25' },
  Nightstand: { left: '86.25%', top: '23%', scale: '0.28' },
  Painting: { left: '7%', top: '20%', scale: '0.20' },
  Sofa: { left: '80%', top: '50%', scale: '0.3' },
  Stove: { left: '50%', top: '23%', scale: '2', width: '8%', height: '10%' },
  Table: { left: '30%', top: '30%', scale: '0.20' },
  Wardrobe: { left: '47.5%', top: '-10%', scale: '0.20' },
  Window: { left: '-17%', top: '0%', scale: '0.20', zIndex: 5 },
};

function Room() {
  const navigate = useNavigate();
  const { userId } = useParams();

  {/*Состояния компонентов нашей страницы:
  user - данные владельца дома
  tasks - список задач на сегодня
  showCabinet - показывать/скрывать личный кабинет (окно)
  cabinetSection - активный раздел кабинета ('tasks', 'shop', 'visitors')
  plans - выбранные пользователем задачи
  shopItems - товары в магазине
  isLoading - состояние загрузки 
  showTaskModal - показывать/скрывать модалку выбора задач
  availableTasks - доступные для добавления задачи
  selectedTaskId - ID выбранной задачи для добавления
  myFurniture -  Мебель пользователя
  searchQuery - запрос дял поиска соседей
  visitedUser - сосед у которого мы в гостях
  visitedFurniture - мебель соседа
  isLoadingRoom - состояние загрузки комнаты соседа
 */}
  const [user, setUser] = useState(null);
  const [tasks, setTasks] = useState([]);
  const [showCabinet, OpenCabinet] = useState(false);
  const [cabinetSection, CabinetButtons] = useState('tasks');
  const [plans, setPlans] = useState([]);
  const [shopItems, Shop] = useState([]);
  const [isLoading, Loading] = useState(true);
  const [showTaskModal, TaskWindow] = useState(false);
  const [availableTasks, TasksMy] = useState([]);
  const [selectedTaskId, TasksIdMy] = useState(null);
  const [myFurniture, FurnitureMy] = useState([]);
  const [searchQuery, FindNeighbour] = useState('');
  const [searchResults, MyNeighbour] = useState([]);
  const [visitedUser, Neighbour] = useState(null);
  const [visitedFurniture, FurnitureNeighbour] = useState([]);
  const [isLoadingRoom, LoadingNeighbour] = useState(false);

  useEffect(() => {
    checkAuth();
    if (userId && userId !== 'me') {
      loadNeighbourRoom(userId);
    }
  }, [userId]);

  const checkAuth = async () => {
    const token = localStorage.getItem('token');

    if (!token) {
      console.log('Нет токена');
      localStorage.clear();
      navigate('/login');
      return;
    }

    try {
      const profileRes = await fetch(API_URL + "/profile", {
        headers: { 'Authorization': "Bearer " + token }
      });

      if (profileRes.ok) {
        const userData = await profileRes.json();
        setUser({
          id: userData.user_id,
          username: userData.username,
          coins: userData.coins
        });
      }
      await LoadMe();

    } catch (error) {
      localStorage.clear();
      navigate('/login');
    } finally {
      Loading(false);
    }
  };

  const loadNeighbourRoom = async (id) => {
    LoadingNeighbour(true);
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_URL}/rooms/${id}`, {
        headers: { 'Authorization': "Bearer " + token }
      });

      if (response.ok) {
        const data = await response.json();
        Neighbour(data.owner);
        FurnitureNeighbour(data.furniture);
      }
    } catch (error) {
      alert('Не удалось загрузить комнату');
    } finally {
      LoadingNeighbour(false);
    }
  };

  {/*
   Получаем все данные, связанные с владельцем дома:
   1. Все его задачи на сегодня;
   2. Всю его мебель;
   3. Все его задачи в целом; 
   4. Всю его мебель;
   5. Его данные.
   Promise.all() для параллельной загрузки
  */}
  const LoadMe = async () => {
    const token = localStorage.getItem('token');

    try {
      const [todayTasksRes, shopRes, catalogRes, myFurnitureRes, profileRes] = await Promise.all([
        fetch(API_URL + "/tasks", { headers: { 'Authorization': "Bearer " + token } }),
        fetch(API_URL + "/furniture/catalog", { headers: { 'Authorization': "Bearer " + token } }),
        fetch(API_URL + "/tasks/catalog", { headers: { 'Authorization': "Bearer " + token } }),
        fetch(API_URL + "/furniture/my", { headers: { 'Authorization': "Bearer " + token } }),
        fetch(API_URL + "/profile", { headers: { 'Authorization': "Bearer " + token } }),
      ]);
      //1. Задачи на сегодня
      if (todayTasksRes.ok) {
        const todayData = await todayTasksRes.json();
        setTasks(todayData.tasks.map(task => ({
          id: task.id,
          text: task.description,
          completed: task.completed || false,
          userTaskId: task.user_task_id,
          streak: task.streak_days || 0
        })));
      }
      //2. Магазин мебели
      if (shopRes.ok) {
        const shopData = await shopRes.json();
        if (shopData.catalog) {
          Shop(shopData.catalog.map(item => ({
            id: item.id,
            name: item.name,
            price: item.price
          })));
        }
      }

      //3. Список задач
      if (catalogRes.ok) {
        const catalogData = await catalogRes.json();
        if (catalogData.catalog) {
          const selectedTasks = catalogData.catalog.filter(task => task.is_selected);

          setPlans(selectedTasks.map(task => ({
            id: task.id,
            name: task.description,
            isActive: task.is_active
          })));
        }
      }

      //4. Мебель хозяина
      if (myFurnitureRes.ok) {
        const myFurnitureData = await myFurnitureRes.json();
        if (myFurnitureData.furniture) {
          FurnitureMy(myFurnitureData.furniture.map(item => ({
            id: item.id,
            name: item.name,
            level: item.level,
            templateId: item.template_id
          })));
        }
      }

      //5. Данные о хозяине
      if (profileRes.ok) {
        const userData = await profileRes.json();
        setUser({
          id: userData.user_id,
          username: userData.username,
          coins: userData.coins
        });
      }

    } catch (error) {
    }
  };

  const toggleTask = (taskId) => {
    setTasks(tasks.map(task =>
      task.id === taskId ? { ...task, completed: !task.completed } : task
    ));
  };

  const submitTasks = async () => {
    const completedTasks = tasks.filter(task => task.completed);
    if (completedTasks.length === 0) {
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(API_URL + "/tasks/submit", {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': "Bearer " + token
        },
        body: JSON.stringify({
          task_ids: completedTasks.map(task => task.id)
        }),
      });

      const data = await response.json();
      if (response.ok) {
        alert("Выполнено задач: " + data.tasks_completed + " Получено: " + data.coins_earned + " монет!");
        await LoadMe();
      }
    } catch (error) {
    }
  };

  const addPlan = async () => {
    try {
      const token = localStorage.getItem('token');

      const response = await fetch(API_URL + "/tasks/catalog", {
        headers: { 'Authorization': "Bearer " + token }
      });

      if (response.ok) {
        const data = await response.json();

        if (data.catalog && data.catalog.length > 0) {
          const tasks = data.catalog.filter(task => !task.is_selected);

          if (tasks.length === 0) {
            alert('Вы уже добавили все доступные задачи!');
            return;
          }

          TasksMy(tasks);
          TasksIdMy(null);
          TaskWindow(true);
        } else {
          alert('Нет доступных задач для добавления');
        }
      } else {
        alert('Ошибка загрузки списка задач');
      }
    } catch (error) {
    }
  };

  const handleSelectTask = (taskId) => {
    TasksIdMy(taskId);
  };

  const handleConfirmSelection = async () => {
    const token = localStorage.getItem('token');

    try {
      const selectedTask = availableTasks.find(task => task.id === selectedTaskId);

      const response = await fetch(API_URL + "/tasks/select", {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': "Bearer " + token
        },
        body: JSON.stringify({
          template_id: selectedTaskId
        })
      });

      if (response.ok) {
        alert("Задача " + selectedTask.description + " добавлена в ваш план!");
        TaskWindow(false);
        await LoadMe();
      } else {
        const errorData = await response.json();
        alert(errorData.error || 'Ошибка добавления задачи');
      }
    } catch (error) {
    }
  };

  const buyItem = async (item) => {
    const token = localStorage.getItem('token');
    if (!user || user.coins < item.price) {
      alert('Недостаточно монет!');
      return;
    }

    try {
      const response = await fetch(API_URL + "/furniture/buy", {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': "Bearer " + token
        },
        body: JSON.stringify({ template_id: item.id }),
      });
      const data = await response.json();
      if (response.ok) {
        await LoadMe();
      }
    } catch (error) {
    }
  };

  const searchAll = async () => {
    if (!searchQuery.trim()) {
      MyNeighbour([]);
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_URL}/search?q=${encodeURIComponent(searchQuery)}`, {
        headers: { 'Authorization': "Bearer " + token }
      });

      if (response.ok) {
        const data = await response.json();
        MyNeighbour(data.users || []);
      }
    } catch (error) {
      console.error('Search error:', error);
    }
  };

  const VisitNeighbour = (userId) => {
    navigate(`/room/${userId}`);
    OpenCabinet(false);
  };

  const returnHome = () => {
    navigate('/room/me');
    Neighbour(null);
    FurnitureNeighbour([]);
  };

  const logout = () => {
    localStorage.clear();
    navigate('/login');
  };

  if (isLoading) {
    return (
      <div className="loading-screen">
        <div className="loading-text">Загрузка...</div>
      </div>
    );
  }
  const currentFurniture = userId && userId !== 'me' ? visitedFurniture : myFurniture;
  const completedCount = tasks.filter(t => t.completed).length;

  return (
    <div className="room-page">
      <nav className="room-navbar">
        <div className="navbar-content">
          <h1 className="navbar-title">
            {userId && userId !== 'me' ? 'В гостях' : 'Дом'}
          </h1>
          <div className="navbar-buttons">
            {userId && userId !== 'me' && (
              <button onClick={returnHome} className="navbar-button pixel-button">
                Вернуться домой
              </button>
            )}
            <button onClick={() => OpenCabinet(true)} className="navbar-button pixel-button">
              Личный кабинет
            </button>
            <button onClick={logout} className="navbar-button pixel-button">
              Выйти
            </button>
          </div>
        </div>
      </nav>

      {showCabinet && (
        <div className="modal-overlay" onClick={() => OpenCabinet(false)}>
          <div className="personal-cabinet-modal" onClick={(e) => e.stopPropagation()}>
            <button
              onClick={() => OpenCabinet(false)}
              className="close-modal-button"
              style={{
                position: 'absolute',
                top: '10px',
                right: '10px',
                background: 'transparent',
                border: 'none',
                fontSize: '24px',
                cursor: 'pointer',
                color: '#fff'
              }}
            >
              ✕
            </button>
            <div className="cabinet-content">
              <div className="cabinet-left">
                {cabinetSection === 'tasks' && (
                  <div className="tasks-section-cabinet">
                    <h3>Мой план</h3>

                    {plans.length === 0 ? (
                      <div className="empty-plans">
                      </div>
                    ) : (
                      <ul className="plans-list">
                        {plans.map(plan => (
                          <li key={plan.id} className="plan-item">
                            <div className="plan-name">{plan.name}</div>
                          </li>
                        ))}
                      </ul>
                    )}

                    <button onClick={addPlan} className="add-plan-button pixel-button">
                      Добавить задачу в план
                    </button>
                  </div>
                )}

                {cabinetSection === 'shop' && (
                  <div className="shop-section-cabinet">
                    <h3>Магазин</h3>
                    <ul className="shop-items-list">
                      {shopItems.map(item => {
                        const canBuy = user && user.coins >= item.price;
                        return (
                          <li key={item.id} className="shop-item">
                            <div className="shop-item-name">{item.name}</div>
                            <div className="shop-item-price">{item.price} монет</div>
                            <button
                              onClick={() => buyItem(item)}
                              className={`buy-button pixel-button ${!canBuy ? 'disabled' : ''}`}
                              disabled={!canBuy}
                            >
                              {canBuy ? 'Купить' : 'Недостаточно'}
                            </button>
                          </li>
                        );
                      })}
                    </ul>
                  </div>
                )}

                {cabinetSection === 'visitors' && (
                  <div className="visitors-section-cabinet">
                    <h3>В гости</h3>

                    <div className="search-container">
                      <div className="search-box">
                        <input
                          type="text"
                          placeholder="Поиск соседей..."
                          className="search-input"
                          value={searchQuery}
                          onChange={(e) => FindNeighbour(e.target.value)}
                          onKeyUp={(e) => {
                            if (e.key === 'Enter') searchAll();
                          }}
                        />
                        <button onClick={searchAll} className="search-button">
                          Найти
                        </button>
                      </div>
                    </div>

                    <div className="visitors-list">
                      {searchResults.length === 0 ? (
                        <p className="no-visitors">
                          {searchQuery ? 'Пользователи не найдены' : 'Введите имя для поиска'}
                        </p>
                      ) : (
                        <ul className="users-list">
                          {searchResults.map(user => (
                            <li key={user.id} className="user-item">
                              <div className="user-name">{user.username}</div>
                              <button
                                onClick={() => VisitNeighbour(user.id)}
                                className="visit-button pixel-button"
                              >
                                В гости
                              </button>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                  </div>
                )}
              </div>

              <div className="cabinet-right" style={{ backgroundImage: 'url(/Images/1.jpg)' }}>
                <div className="profile-section">
                  <div className="avatar-container">
                    <img src="/Images/avatar.jpg" className="avatar-image" />
                    <img src="/Images/ramka.png" className="avatar-frame" />
                  </div>
                  <div className="username">{user.username}</div>
                </div>

                <div className="cabinet-buttons">
                  <button onClick={() => CabinetButtons('tasks')} className={`cabinet-button pixel-button ${cabinetSection === 'tasks' ? 'active' : ''}`} >
                    <img src="/Images/palochka.png" className="button-icon" />
                    Мои задачи
                  </button>
                  <button onClick={() => CabinetButtons('shop')} className={`cabinet-button pixel-button ${cabinetSection === 'shop' ? 'active' : ''}`}>
                    <img src="/Images/myhome.png" className="button-icon" />
                    Магазин
                  </button>
                  <button onClick={() => CabinetButtons('visitors')} className={`cabinet-button pixel-button ${cabinetSection === 'visitors' ? 'active' : ''}`}>
                    <img src="/Images/loshad.png" className="button-icon" />
                    В гости
                  </button>
                </div>

                <div className="cabinet-currency">
                  <div className="currency-title">Монеты:</div>
                  <div className="currency-amount-cabinet">{user.coins}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Модальное окно выбора задач */}
      {showTaskModal && (
        <div className="task-modal-overlay" onClick={() => TaskWindow(false)}>
          <div className="task-modal" onClick={(e) => e.stopPropagation()}>
            <div className="task-modal-header">
              <h3>Выберите задачу для добавления</h3>
              <button
                onClick={() => TaskWindow(false)}
                className="close-task-modal"
              >
                ✕
              </button>
            </div>

            <div className="task-modal-content">
              {availableTasks.length === 0 ? (
                <p className="no-tasks-msg">Нет доступных задач</p>
              ) : (
                <ul className="task-selection-list">
                  {availableTasks.map(task => (
                    <li
                      key={task.id}
                      className={`task-selection-item ${selectedTaskId === task.id ? 'selected' : ''}`}
                    >
                      <div className="task-info">
                        <div className="task-title">{task.description}</div>
                        <div className="task-reward">Награда: {task.base_reward} монета штук</div>
                      </div>
                      <button
                        onClick={() => handleSelectTask(task.id)}
                        className={`select-task-button ${selectedTaskId === task.id ? 'selected' : ''}`}
                      >
                        {selectedTaskId === task.id ? 'Выбрано' : 'Выбрать'}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>

            <div className="task-modal-footer">
              <button
                onClick={() => TaskWindow(false)}
                className="cancel-button pixel-button"
              >
                Отмена
              </button>
              <button
                onClick={handleConfirmSelection}
                className="confirm-button pixel-button"
              >
                Добавить задачу
              </button>
            </div>
          </div>
        </div>
      )}

      {isLoadingRoom ? (
        <div className="loading-screen">
          <div className="loading-text">Загрузка комнаты...</div>
        </div>
      ) : (
        <div className="room-content">
          {(!userId || userId === 'me') && (
            <div className="tasks-section">
              <h2 className="section-title">Задачи на сегодня</h2>
              <ul className="tasks-list">
                {tasks.map(task => (
                  <li key={task.id} className="task-item">
                    <label className="task-label">
                      <input type="checkbox" checked={task.completed} onChange={() => toggleTask(task.id)} />
                      <span className={`task-text ${task.completed ? 'completed' : ''}`}>{task.text}</span>
                    </label>
                  </li>
                ))}
              </ul>
              <button onClick={submitTasks} className="submit-button pixel-button">Отправить</button>
            </div>
          )}

          <div className="room-section">
            <div className="room-container">
              <div className="room">
                <img src="/Images/clean_room.png" alt="Комната" className="room-background" />

                {currentFurniture && currentFurniture.map(item => {
                  const furnitureType = item.name.split(' ')[0];
                  const pos = furnituree_pos[furnitureType];
                  if (!pos) return null;

                  return (
                    <div
                      key={`${furnitureType}-${item.level}-${item.id}`}
                      style={{
                        position: 'absolute',
                        left: pos.left,
                        top: pos.top,
                        zIndex: pos.zIndex,
                        width: pos.width,
                        height: pos.height,
                      }}
                    >
                      <img
                        src={`/Images/furniture/${furnitureType}${item.level}.png`}
                        alt={furnitureType}
                        style={{
                          transform: `scale(${pos.scale})`
                        }}
                      />
                    </div>
                  );
                })}
              </div>
            </div>
          </div>

          {(!userId || userId === 'me') && (
            <div className="info-section">
              <div className="info-card">
                <h3 className="info-title">Монеты</h3>
                <div className="currency-amount">{user.coins}</div>
              </div>

              <div className="info-card">
                <h3 className="info-title">Прогресс на сегодня</h3>
                <div className="progress-bar">
                  <div className="progress-fill" style={{ width: (tasks.length > 0 ? (completedCount / tasks.length) * 100 : 0) + "%" }}></div>
                </div>
                <p className="info-text">Выполнено: {completedCount} из {tasks.length}</p>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default Room;