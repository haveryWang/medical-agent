export default function Select({ label, value, options, onChange }) {
  return (
    <label className="select-wrap">
      <span>{label}:</span>
      <select value={value} onChange={(e) => onChange(e.target.value)}>
        {options.map((option) => <option key={option || 'all'} value={option}>{option || `全部${label}`}</option>)}
      </select>
    </label>
  );
}
