import { GithubOutlined } from '@ant-design/icons';
import { DefaultFooter } from '@ant-design/pro-components';

const Footer: React.FC = () => {
  const defaultMessage = 'DBMETA数据库数据管理平台， Power by dbmeta.com，Version:1.0';

  const currentYear = new Date().getFullYear();

  return (
    <DefaultFooter
      copyright={`${currentYear} ${defaultMessage}`}
      links={[
        {
          key: 'Lepus-cc',
          title: 'DBMETA官方网站',
          href: 'https://www.dbmeta.com',
          blankTarget: true,
        },
        {
          key: 'github',
          title: <GithubOutlined />,
          href: 'https://gitee.com/lepus-group',
          blankTarget: true,
        },
        {
          key: 'discuss-lepus',
          title: 'Lepus交流社区',
          href: 'https://discuss.lepus.cc',
          blankTarget: true,
        },
      ]}
    />
  );
};

export default Footer;
