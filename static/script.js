let vizInstance; // Глобальная переменная для хранения экземпляра Viz.js
let graphData; // Глобальная переменная для хранения данных графа

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

    // Получаем данные графа с сервера
    const graphResponse = await fetch('/graph');
    if (!graphResponse.ok) {
      throw new Error('Не удалось получить данные графа.');
    }

    graphData = await graphResponse.json(); // Сохраняем данные графа
    renderGraph(); // Рисуем граф после загрузки данных
  } catch (error) {
    console.error('Ошибка:', error);
    alert(error.message || 'Не удалось загрузить файл или построить граф.');
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
    const label = events; // Показываем только количество событий
    dot += `  "${edge.data.from}" -> "${edge.data.to}" [label="${label}"];\n`;
  });

  dot += '}';
  return dot;
}

// Отрисовка графа
async function renderGraph() {
  try {
    if (!graphData) {
      throw new Error('Граф еще не загружен. Загрузите CSV-файл.');
    }

    const powerSlider = document.getElementById('power-slider');
    const powerValue = parseInt(powerSlider.value); // Текущее значение ползунка (0–100%)

    // Получаем диапазон мощности ребер
    const counts = graphData.edges.map(edge => edge.data.count);
    const min = Math.min(...counts);
    const max = Math.max(...counts);

    // Вычисляем пороговое значение мощности
    const threshold = min + ((max - min) * (100 - powerValue)) / 100;

    // Фильтрация ребер по мощности
    const filteredEdges = graphData.edges.filter(edge => edge.data.count >= threshold);

    // Создаем новый объект данных с отфильтрованными ребрами
    const filteredData = {
      nodes: graphData.nodes, // Узлы остаются без изменений
      edges: filteredEdges, // Только отфильтрованные ребра
    };

    const dot = convertToDot(filteredData); // Преобразование данных в формат DOT

    if (!vizInstance) {
      vizInstance = new Viz({
        workerURL: "/js/full.render.js", // Локальный путь
      });
    }

    // Рендеринг DOT в SVG
    const svg = await vizInstance.renderString(dot);
    const graphContainer = document.getElementById('graph');

    // Очищаем контейнер перед новой отрисовкой
    graphContainer.innerHTML = '';

    // Вставляем SVG в DOM
    graphContainer.innerHTML = svg;

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

    // Получаем реальные размеры SVG
    const svgWidth = svgElement.width.baseVal.value || svgElement.clientWidth;
    const svgHeight = svgElement.height.baseVal.value || svgElement.clientHeight;

    // Устанавливаем размеры canvas с учетом DPR (Device Pixel Ratio)
    const dpr = window.devicePixelRatio || 1; // Плотность пикселей устройства
    canvas.width = svgWidth * dpr;
    canvas.height = svgHeight * dpr;

    // Масштабируем контекст canvas для повышения качества
    ctx.scale(dpr, dpr);

    // Создаем изображение из SVG
    const img = new Image();
    img.onload = () => {
      ctx.drawImage(img, 0, 0, svgWidth, svgHeight);

      // Конвертируем canvas в PNG
      canvas.toBlob(
        (blob) => {
          const pngUrl = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = pngUrl;
          a.download = 'graph.png';
          document.body.appendChild(a);
          a.click();
          a.remove();
          URL.revokeObjectURL(pngUrl);
        },
        'image/png',
        1.0 // Настройка качества PNG (1.0 — максимальное качество)
      );
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
  const powerSlider = document.getElementById('power-slider');
  const powerValue = document.getElementById('power-value');

  if (!fileInput || !uploadBtn || !downloadBtn || !powerSlider || !powerValue) {
    console.error('Один или несколько элементов DOM не найдены.');
    return;
  }

  // Клик на кнопку "Загрузить файл"
  uploadBtn.addEventListener('click', () => {
    fileInput.click(); // Программно вызываем выбор файла
  });

  // Обработка выбора файла
  fileInput.addEventListener('change', () => {
    const file = fileInput.files[0];
    if (file) {
      uploadFile(file); // Автоматическая загрузка файла при выборе
      downloadBtn.disabled = false; // Активируем кнопку скачивания
    } else {
      alert('Выберите файл для загрузки.');
    }
  });

  // Клик на кнопку "Скачать PNG"
  downloadBtn.addEventListener('click', downloadPNG);

  // Изменение значения ползунка
  powerSlider.addEventListener('input', () => {
    powerValue.textContent = `${powerSlider.value}%`; // Обновляем отображаемое значение
    renderGraph(); // Перерисовываем граф при изменении ползунка
  });
});