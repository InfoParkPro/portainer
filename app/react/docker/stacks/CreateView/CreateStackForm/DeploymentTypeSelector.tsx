import DockerCompose from '@/assets/ico/vendor/docker-compose.svg?c';
import DockerIcon from '@/assets/ico/vendor/docker-icon.svg?c';

import { BoxSelector } from '@@/BoxSelector';
import { BoxSelectorOption } from '@@/BoxSelector/types';
import { FormSection } from '@@/form-components/FormSection';

import { DockerDeploymentType } from './types';

const deploymentTypeOptions: Array<BoxSelectorOption<DockerDeploymentType>> = [
  {
    id: 'deployment_swarm',
    icon: DockerIcon,
    iconType: 'logo',
    label: 'Swarm stack',
    description: 'Deploy as Docker Swarm services',
    value: 'swarm',
  },
  {
    id: 'deployment_compose',
    icon: DockerCompose,
    iconType: 'logo',
    label: 'Compose stack',
    description: 'Deploy as Docker Compose containers',
    value: 'standalone',
  },
];

export function DeploymentTypeSelector({
  value,
  onChange,
}: {
  value: DockerDeploymentType;
  onChange(value: DockerDeploymentType): void;
}) {
  return (
    <FormSection title="Deployment type">
      <BoxSelector
        radioName="deployment-type"
        value={value}
        onChange={onChange}
        options={deploymentTypeOptions}
        slim
      />
    </FormSection>
  );
}
