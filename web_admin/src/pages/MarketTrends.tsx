import React, { useEffect, useState } from 'react';
import { Table, Tag, Card, Popconfirm, Button, message, Modal, Form, Input, InputNumber, Select, Switch, Tooltip } from 'antd';
import { InfoCircleOutlined } from '@ant-design/icons';
import { getMarketTrends, updateMarketTrend, deleteMarketTrend, type MarketTrend } from '../api';

const MarketTrends: React.FC = () => {
  const [data, setData] = useState<MarketTrend[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingTrend, setEditingTrend] = useState<MarketTrend | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadData(1, 20);
  }, []);

  const loadData = async (page: number, pageSize: number) => {
    setLoading(true);
    try {
      const res = await getMarketTrends(page, pageSize);
      setData(res.trends || []);
      setPagination(prev => ({
        ...prev,
        current: page,
        pageSize: pageSize,
        total: res.total,
      }));
      setSelectedRowKeys([]); // clear selection
    } catch (e) {
      console.error(e);
      message.error('加载失败');
    } finally {
      setLoading(false);
    }
  };

  const handleTableChange = (newPagination: any) => {
    loadData(newPagination.current, newPagination.pageSize);
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteMarketTrend(id);
      message.success('删除成功');
      loadData(pagination.current, pagination.pageSize);
    } catch (e) {
      console.error(e);
      message.error('删除失败');
    }
  };

  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要删除的趋势');
      return;
    }

    try {
      setLoading(true);
      await Promise.all(selectedRowKeys.map(key => deleteMarketTrend(key as number)));
      message.success(`成功删除 ${selectedRowKeys.length} 条趋势`);
      loadData(pagination.current, pagination.pageSize);
    } catch (e) {
      console.error(e);
      message.error('批量删除过程中出现错误');
      loadData(pagination.current, pagination.pageSize);
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (record: MarketTrend) => {
    setEditingTrend(record);
    form.setFieldsValue({
      ...record,
      related_sectors: record.related_sectors || [], // ensure array
    });
    setIsModalVisible(true);
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      if (editingTrend) {
        await updateMarketTrend({ ...editingTrend, ...values });
        message.success('更新成功');
        setIsModalVisible(false);
        loadData(pagination.current, pagination.pageSize);
      }
    } catch (e) {
      console.error(e);
      message.error('操作失败');
    }
  };

  const columns = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (text: string, record: MarketTrend) => (
        <div>
          <a href={record.original_url} target="_blank" rel="noreferrer" style={{ fontWeight: 'bold' }}>{text}</a>
          <div style={{ fontSize: 12, color: '#888' }}>{record.source}</div>
        </div>
      ),
    },
    {
      title: '影响类型',
      dataIndex: 'impact_type',
      key: 'impact_type',
      render: (type: string) => (
        <Tag color={type === 'policy_long_term' ? 'purple' : 'blue'}>
          {type === 'policy_long_term' ? '长期政策' : '短期消息'}
        </Tag>
      ),
    },
    {
      title: '影响范围',
      dataIndex: 'impact_scope',
      key: 'impact_scope',
      render: (scope: string) => (
        <Tag color={scope === 'broad' ? 'magenta' : 'geekblue'}>
          {scope === 'broad' ? '广泛' : '特定'}
        </Tag>
      ),
    },
    {
      title: (
        <span>
          相关性
          <Tooltip title="金融相关度评分 (0-10)：衡量该信息是否属于财经范畴。高分表示重要财经新闻，低分表示可能是噪音或八卦。">
            <InfoCircleOutlined style={{ marginLeft: 4 }} />
          </Tooltip>
        </span>
      ),
      dataIndex: 'financial_relevance',
      key: 'financial_relevance',
      render: (val: number) => (
         <Tag color={val >= 8 ? 'green' : val >= 5 ? 'orange' : 'default'}>{val}</Tag>
      ),
    },
    {
      title: '情感评分',
      dataIndex: 'sentiment_score',
      key: 'sentiment_score',
      render: (val: number) => val.toFixed(2),
    },
    {
      title: '相关板块',
      dataIndex: 'related_sectors',
      key: 'related_sectors',
      render: (sectors: string[]) => (
        <>
          {sectors && sectors.length > 0 ? (
            sectors.map((tag) => (
              <Tag color="blue" key={tag}>
                {tag}
              </Tag>
            ))
          ) : (
            <span style={{ color: '#ccc' }}>无特定板块</span>
          )}
        </>
      ),
    },
    {
      title: '相关股票',
      dataIndex: 'related_stocks',
      key: 'related_stocks',
      render: (stocks: string[]) => (
        <>
          {stocks && stocks.length > 0 ? (
            stocks.map((tag) => {
              // Handle "Name|Code" format
              let display = tag;
              let code = "";
              if (tag.includes('|')) {
                const parts = tag.split('|');
                display = parts[0];
                code = parts[1];
              }
              return (
                <Tooltip title={code || tag} key={tag}>
                  <Tag color="cyan">
                    {display}
                  </Tag>
                </Tooltip>
              );
            })
          ) : (
            <span style={{ color: '#ccc' }}>无特定股票</span>
          )}
        </>
      ),
    },
    {
      title: '权重',
      dataIndex: 'weight',
      key: 'weight',
    },
    {
      title: '有效性',
      dataIndex: 'is_still_valid',
      key: 'is_still_valid',
      render: (valid: boolean) => (
        <Tag color={valid ? 'success' : 'error'}>{valid ? '有效' : '失效'}</Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (t: string) => new Date(t).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: MarketTrend) => (
        <span>
          <Button type="link" onClick={() => handleEdit(record)}>编辑</Button>
          <Popconfirm title="确定要删除吗？" onConfirm={() => handleDelete(record.id)} okText="确定" cancelText="取消">
            <Button type="link" danger style={{ marginLeft: 8 }}>删除</Button>
          </Popconfirm>
        </span>
      ),
    },
  ];

  return (
    <Card 
      title="市场趋势管理"
      extra={
        selectedRowKeys.length > 0 && (
          <Popconfirm 
            title={`确定要删除选中的 ${selectedRowKeys.length} 条趋势吗？`} 
            onConfirm={handleBatchDelete} 
            okText="确定" 
            cancelText="取消"
          >
            <Button type="primary" danger>批量删除</Button>
          </Popconfirm>
        )
      }
    >
      <Table 
        dataSource={data} 
        columns={columns} 
        rowKey="id" 
        loading={loading}
        rowSelection={{
          selectedRowKeys,
          onChange: (newSelectedRowKeys) => setSelectedRowKeys(newSelectedRowKeys),
        }}
        pagination={{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizeChanger: true,
          defaultPageSize: 20,
          pageSizeOptions: ['10', '20', '50', '100'],
        }}
        onChange={handleTableChange}
      />

      <Modal
        title="编辑市场趋势"
        open={isModalVisible}
        onOk={handleModalOk}
        onCancel={() => setIsModalVisible(false)}
        width={800}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="summary" label="摘要">
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.Item name="impact_analysis" label="影响分析">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="related_stocks" label="相关股票">
            <Select mode="tags" style={{ width: '100%' }} placeholder="输入 '股票名称|代码' 或仅代码后回车" />
          </Form.Item>
          <Form.Item name="financial_relevance" label="金融相关性 (0-10)">
            <InputNumber min={0} max={10} />
          </Form.Item>
          <Form.Item name="weight" label="权重">
            <InputNumber step={0.1} />
          </Form.Item>
          <Form.Item name="impact_type" label="影响类型">
            <Select>
              <Select.Option value="policy_long_term">长期政策</Select.Option>
              <Select.Option value="short_term_news">短期消息</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="impact_scope" label="影响范围">
            <Select>
              <Select.Option value="specific">特定</Select.Option>
              <Select.Option value="broad">广泛</Select.Option>
            </Select>
          </Form.Item>
           <Form.Item name="is_still_valid" label="是否有效" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default MarketTrends;
