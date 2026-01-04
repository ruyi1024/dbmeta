// 类型声明文件，解决模块导入问题
declare module 'antd' {
  const antd: any;
  export = antd;
  export * from 'antd';
}

declare module '@ant-design/icons' {
  const icons: any;
  export = icons;
  export * from '@ant-design/icons';
}

declare module '@ant-design/pro-layout' {
  const proLayout: any;
  export = proLayout;
  export * from '@ant-design/pro-layout';
}

declare module 'moment' {
  interface Moment {
    format(format?: string): string;
  }
  
  function moment(input?: any): Moment;
  export = moment;
}

declare module 'umi' {
  export const history: {
    push: (path: string) => void;
    replace: (path: string) => void;
    goBack: () => void;
  };
}

// 全局类型声明
declare global {
  namespace JSX {
    interface IntrinsicElements {
      [elemName: string]: any;
    }
  }
}
