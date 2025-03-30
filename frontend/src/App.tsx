import React from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Home from '@/pages/Home';
import GraphView from '@/pages/GraphView';


const App: React.FC = () => {
  return (
    <Router>
      <Routes>
        <Route path='/' element={<Home />} />
        <Route path='/graph' element={<GraphView />} />
      </Routes>
    </Router>
  )
}

export default App
