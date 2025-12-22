import * as _ from 'lodash';

export const renderTextTemplate = (template: string, vars: Record<string, any>): string => {
  if (!template?.trim()) return template;

  const patterns = template.match(/\{\{[^\}]+\}\}/g);

  return !patterns ? template : patterns.reduce((resultText, pattern) => {
    const fieldPath = pattern.replace(/[\{\}]/g, '');
    const value = _.get(vars, fieldPath);
    return (resultText as any).replaceAll(pattern, value);
  }, template);
}
