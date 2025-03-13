let currentMode = 'events';
let vizInstance;
let isGraphBuilt = false; // Флаг для отслеживания состояния графа

// Функция для отправки файла на сервер
async function uploadFile(file) {
  const formData = new FormData();
  formData.append('file', file);

  try {
    const response = await fetch('/upload', {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error('Ошибка загрузки файла');
    }

    alert('Файл успешно загружен!');
    isGraphBuilt = true;
    document.getElementById('download-btn').disabled = false; // Активируем кнопку скачивания
    renderGraph(); // После загрузки файла рисуем граф
  } catch (error) {
    console.error('Ошибка:', error);
    alert('Не удалось загрузить файл');
  }
}

// Преобразование данных в формат DOT
function convertToDot(data) {
  let dot = 'digraph G {\n';
  dot += '  rankdir=LR;\n'; // Направление графа слева направо
  dot += '  node [shape=rect style=filled];\n'; // Стиль узлов
  dot += '  edge [fontsize=12];\n'; // Стиль ребер

  // Добавление узлов
  data.nodes.forEach(node => {
    const color = node.data.color || '#add8e6'; // Цвет узла
    const label = `${node.data.label} (${node.data.count})`; // Метка узла
    dot += `  "${node.data.id}" [label="${label}" fillcolor="${color}"];\n`;
  });

  // Добавление ребер
  data.edges.forEach(edge => {
    const [events, time] = edge.data.label.split('\n'); // Разделение метки на события и время
    const label = currentMode === 'events' ? events : time; // Выбор метки в зависимости от режима
    const style = edge.data.style === "dashed" ? " [style=dashed]" : ""; // Проверяем стиль линии
    dot += `  "${edge.data.from}" -> "${edge.data.to}" [label="${label}"${style}];\n`;
  });

  dot += '}';
  return dot;
}

// Отрисовка графа
async function renderGraph() {
  try {
    const response = await fetch('/graph');
    if (!response.ok) {
      throw new Error('Граф еще не построен. Загрузите CSV-файл.');
    }

    const graphData = await response.json();
    console.log('Graph Data:', graphData); // Логирование данных

    const dot = convertToDot(graphData); // Преобразование данных в формат DOT
    console.log('DOT String:', dot); // Логирование DOT-строки

    if (!vizInstance) {
      vizInstance = new Viz({
        workerURL: "/js/full.render.js", // Локальный путь
      });
    }

    // Рендеринг DOT в SVG
    const svg = await vizInstance.renderString(dot);
    const graphContainer = document.getElementById('graph');
    graphContainer.innerHTML = svg; // Вставка SVG в DOM

    // Инициализация Panzoom
    const panzoomElement = graphContainer.querySelector('svg');
    if (panzoomElement) {
      const panzoom = Panzoom(panzoomElement, {
        maxScale: 5, // Максимальное масштабирование
        minScale: 0.5, // Минимальное масштабирование
        contain: 'outside', // Удерживать содержимое внутри контейнера
      });

      // Включение зума колесиком мыши
      graphContainer.addEventListener('wheel', (e) => {
        e.preventDefault();
        panzoom.zoomWithWheel(e);
      });

      // Центрирование графа
      panzoom.pan(0, 0);
      panzoom.zoom(1);
    }
  } catch (error) {
    console.error('Ошибка рендеринга графа:', error);
    alert(error.message || 'Не удалось отобразить граф');
  }
}

// Обновление графа при смене режима
function updateGraph() {
  currentMode = document.querySelector('input[name="edge-label"]:checked').value; // Получение текущего режима
  renderGraph(); // Перерисовка графа
}

// Функция для скачивания графа в формате PNG
async function downloadPNG() {
  try {
    const graphContainer = document.getElementById('graph');
    const svgElement = graphContainer.querySelector('svg');

    if (!svgElement) {
      throw new Error('Граф еще не построен. Загрузите CSV-файл.');
    }

    // Клонируем SVG для корректного экспорта
    const clone = svgElement.cloneNode(true);
    const serializer = new XMLSerializer();
    const svgString = serializer.serializeToString(clone);

    // Создаем Blob из SVG
    const blob = new Blob([svgString], { type: 'image/svg+xml;charset=utf-8' });
    const url = URL.createObjectURL(blob);

    // Создаем временный canvas для конвертации SVG в PNG
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    const img = new Image();

    img.onload = () => {
      canvas.width = img.width;
      canvas.height = img.height;
      ctx.drawImage(img, 0, 0);

      // Конвертируем canvas в PNG
      canvas.toBlob((blob) => {
        const pngUrl = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = pngUrl;
        a.download = 'graph.png';
        document.body.appendChild(a);
        a.click();
        a.remove();
        URL.revokeObjectURL(pngUrl);
      }, 'image/png');
    };

    img.src = url;
  } catch (error) {
    console.error('Ошибка скачивания PNG:', error);
    alert(error.message || 'Не удалось скачать PNG');
  }
}

// Инициализация
document.addEventListener('DOMContentLoaded', () => {
  const fileInput = document.getElementById('file-input');
  const uploadBtn = document.getElementById('upload-btn');
  const downloadBtn = document.getElementById('download-btn');

  uploadBtn.addEventListener('click', () => {
    if (!fileInput.files.length) {
      alert('Выберите файл для загрузки');
      return;
    }

    const file = fileInput.files[0];
    uploadFile(file);
  });

  downloadBtn.addEventListener('click', downloadPNG);

  document.querySelectorAll('input[name="edge-label"]').forEach(input => {
    input.addEventListener('change', updateGraph); // Обработчик изменения режима
  });
});