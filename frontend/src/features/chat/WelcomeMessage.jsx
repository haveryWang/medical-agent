export default function WelcomeMessage() {
  return (
    <article className="message assistant">
      <div className="bot-icon">中</div>
      <div className="bubble">
        <p>根据《2型糖尿病防治指南（2023年版）》，2型糖尿病的治疗规范主要包括生活方式干预、血糖监测、药物治疗和定期随访。请上传或选择知识库后开始精准问答。</p>
        <div className="citations">
          <b>引用来源</b>
          <span>1. 2型糖尿病防治指南（2023年版）</span>
          <span>2. 中国2型糖尿病防治指南</span>
          <span>3. 基层医疗机构2型糖尿病管理专家共识</span>
        </div>
      </div>
    </article>
  );
}
