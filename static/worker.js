self.addEventListener('message', async (event) => {
    const { nodes, edges } = event.data; // Предполагается, что данные содержат nodes и edges
  
    if (!Array.isArray(edges)) {
      self.postMessage({ error: "Некорректные данные: edges должен быть массивом" });
      return;
    }
  
    // Фильтруем ребра (пример фильтрации)
    const filteredEdges = edges.filter(edge => edge.data.count > 10);
  
    // Возвращаем обработанные данные
    self.postMessage({ nodes, edges: filteredEdges });
  });