import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { Layout, Menu } from 'antd';
import { AreaChartOutlined, DashboardOutlined } from '@ant-design/icons';
import EvaluationList from './pages/EvaluationList';

const { Header, Content, Footer, Sider } = Layout;

const App: React.FC = () => {
  return (
    <Router>
      <Layout style={{ minHeight: '100vh' }}>
        <Sider collapsible>
          <div style={{ 
            height: 32, 
            margin: 16, 
            background: 'rgba(255, 255, 255, 0.2)', 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            color: 'white',
            fontWeight: 'bold',
            borderRadius: 6
          }}>
            股票管理后台
          </div>
          <Menu 
            theme="dark" 
            defaultSelectedKeys={['1']} 
            mode="inline"
            items={[
              {
                key: '1',
                icon: <AreaChartOutlined />,
                label: <Link to="/evaluations">评测管理</Link>
              }
            ]}
          />
        </Sider>
        <Layout className="site-layout">
          <Header className="site-layout-background" style={{ padding: 0 }} />
          <Content style={{ margin: '0 16px' }}>
            <div style={{ padding: 24, minHeight: 360 }}>
              <Routes>
                <Route path="/evaluations" element={<EvaluationList />} />
                <Route path="/" element={<EvaluationList />} />
              </Routes>
            </div>
          </Content>
          <Footer style={{ textAlign: 'center' }}>股票助手管理后台 ©2026</Footer>
        </Layout>
      </Layout>
    </Router>
  );
};

export default App;
