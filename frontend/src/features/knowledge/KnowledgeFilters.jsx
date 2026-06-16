import { Button, Input, Select, Space } from 'antd';
import { PlusOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons';

export default function KnowledgeFilters({ filters, options, setFilters, onCreate, onReset, onSearch }) {
  const update = (name, value) => setFilters({ ...filters, [name]: value || '' });

  return (
    <div className="knowledge-toolbar">
      <Space wrap size={10}>
        <Input
          allowClear
          className="filter-keyword"
          prefix={<SearchOutlined />}
          placeholder="搜索名称或描述"
          value={filters.keyword}
          onChange={(event) => update('keyword', event.target.value)}
          onPressEnter={onSearch}
        />
        <Select
          allowClear
          showSearch
          placeholder="场景"
          className="filter-select"
          value={filters.scenario || undefined}
          options={options.scenarios.map((value) => ({ label: value, value }))}
          onChange={(value) => update('scenario', value)}
        />
        <Select
          allowClear
          showSearch
          placeholder="标签"
          className="filter-select"
          value={filters.tag || undefined}
          options={options.tags.map((value) => ({ label: value, value }))}
          onChange={(value) => update('tag', value)}
        />
        <Select
          allowClear
          showSearch
          placeholder="部门"
          className="filter-select"
          value={filters.department || undefined}
          options={options.departments.map((value) => ({ label: value, value }))}
          onChange={(value) => update('department', value)}
        />
        <Button type="primary" icon={<SearchOutlined />} onClick={onSearch}>查询</Button>
        <Button icon={<ReloadOutlined />} onClick={onReset}>重置</Button>
      </Space>
      <Button type="primary" icon={<PlusOutlined />} onClick={onCreate}>新建知识库</Button>
    </div>
  );
}
