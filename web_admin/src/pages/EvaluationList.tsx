import React, { useEffect, useState } from 'react';
import { Table, Tag, Card, Popconfirm, Button, message } from 'antd';
import { getEvaluations, deleteEvaluation, type EvaluationRecord } from '../api';

const EvaluationList: React.FC = () => {
  const [data, setData] = useState<EvaluationRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await getEvaluations();
      setData(res.evaluations || []);
      setSelectedRowKeys([]); // clear selection after reload
    } catch (e) {
      console.error(e);
      message.error('加载失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteEvaluation(id);
      message.success('删除成功');
      loadData();
    } catch (e) {
      console.error(e);
      message.error('删除失败');
    }
  };

  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要删除的记录');
      return;
    }

    try {
      setLoading(true);
      // Execute deletions in parallel
      await Promise.all(selectedRowKeys.map(key => deleteEvaluation(key as string)));
      message.success(`成功删除 ${selectedRowKeys.length} 条记录`);
      loadData();
    } catch (e) {
      console.error(e);
      message.error('批量删除过程中出现错误');
      loadData(); // reload to show remaining items
    } finally {
      setLoading(false);
    }
  };

  const statusMap: Record<string, string> = {
    'completed': '已完成',
    'pending': '进行中',
    'failed': '失败',
  };

  const columns = [
    {
      title: '预测日期',
      dataIndex: 'prediction_date',
      key: 'prediction_date',
    },
    {
      title: '股票信息',
      dataIndex: 'stock_code',
      key: 'stock_code',
      render: (val: string, record: EvaluationRecord) => (
        <span>
          {record.stock_name && <span style={{ marginRight: 8, fontWeight: 'bold' }}>{record.stock_name}</span>}
          <span style={{ color: '#888' }}>{val}</span>
        </span>
      ),
    },
    {
      title: '初始价格',
      dataIndex: 'initial_price',
      key: 'initial_price',
      render: (val: number) => val.toFixed(2),
    },
    {
      title: 'T+1 价格',
      dataIndex: 'price_1d',
      key: 'price_1d',
      render: (val: number) => val > 0 ? val.toFixed(2) : '-',
    },
    {
      title: 'T+3 价格',
      dataIndex: 'price_3d',
      key: 'price_3d',
      render: (val: number) => val > 0 ? val.toFixed(2) : '-',
    },
    {
      title: '评分',
      dataIndex: 'score',
      key: 'score',
      render: (val: number) => (
        <Tag color={val > 80 ? 'green' : val > 50 ? 'orange' : 'red'}>
          {val.toFixed(1)}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'completed' ? 'blue' : 'default'}>{statusMap[status] || status}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: EvaluationRecord) => (
        <Popconfirm title="确定要删除吗？" onConfirm={() => handleDelete(record.id)} okText="确定" cancelText="取消">
          <Button type="link" danger>删除</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title="预测评测列表"
      extra={
        selectedRowKeys.length > 0 && (
          <Popconfirm
            title={`确定要删除选中的 ${selectedRowKeys.length} 条记录吗？`}
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
          position: ['bottomRight'],
          showSizeChanger: true,
          defaultPageSize: 10,
          showTotal: (total) => `共 ${total} 条`
        }}
      />
    </Card>
  );
};

export default EvaluationList;
