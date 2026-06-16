import Select from '../../components/Select.jsx';

export default function KnowledgeFilters({ filters, setFilters, onSearch, onUpload }) {
  return (
    <div className="filters">
      <Select label="场景" value={filters.scenario} options={['', '临床诊疗', '药学服务', '医保管理', '护理管理']} onChange={(value) => setFilters({ ...filters, scenario: value })} />
      <Select label="标签" value={filters.tag} options={['', '诊疗', '规范', '医保', '护理']} onChange={(value) => setFilters({ ...filters, tag: value })} />
      <Select label="部门" value={filters.department} options={['', '医务部', '药剂科', '医保办', '护理部']} onChange={(value) => setFilters({ ...filters, department: value })} />
      <input value={filters.keyword} onChange={(e) => setFilters({ ...filters, keyword: e.target.value })} placeholder="搜索知识库名称、描述" />
      <button className="primary small" onClick={onSearch}>查询</button>
      <button className="primary small" onClick={onUpload}>＋ 上传文档</button>
    </div>
  );
}
