import type { DropDownProps } from 'antd/es/dropdown';
import { Dropdown } from 'antd';
import React from 'react';
import classNames from 'classnames';
import styles from './index.less';

export type HeaderDropdownProps = {
  overlayClassName?: string;
  overlay?: React.ReactNode | (() => React.ReactNode) | any;
  menu?: DropDownProps['menu'];
  placement?: 'bottomLeft' | 'bottomRight' | 'topLeft' | 'topCenter' | 'topRight' | 'bottomCenter';
} & Omit<DropDownProps, 'overlay'>;

const HeaderDropdown: React.FC<HeaderDropdownProps> = ({ overlayClassName: cls, overlay, ...restProps }) => {
  // antd5 中 overlay 已废弃，如果提供了 overlay，使用 dropdownRender 兼容
  const dropdownRender = overlay ? () => overlay : undefined;
  return (
    <Dropdown 
      dropdownRender={dropdownRender}
      overlayClassName={classNames(styles.container, cls)} 
      {...restProps} 
    />
  );
};

export default HeaderDropdown;
