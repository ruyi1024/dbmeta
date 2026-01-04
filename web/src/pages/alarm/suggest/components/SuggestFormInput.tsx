import ReactQuill, {Quill} from "react-quill";
import React, {useEffect, useState} from "react";

import 'react-quill/dist/quill.snow.css';
// @ts-ignore
import quillEmoji from 'quill-emoji';
import "quill-emoji/dist/quill-emoji.css";
//注册ToolbarEmoji，将在工具栏出现emoji；注册TextAreaEmoji，将在文本输入框处出现emoji。VideoBlot是我自定义的视频组件，后面会讲，
const { EmojiBlot, ShortNameEmoji, ToolbarEmoji, TextAreaEmoji } = quillEmoji;
Quill.register({
  'formats/emoji': EmojiBlot,
  'modules/emoji-shortname': ShortNameEmoji,
  'modules/emoji-toolbar': ToolbarEmoji,
  'modules/emoji-textarea': TextAreaEmoji,
  // 'modules/ImageExtend': ImageExtend, //拖拽图片扩展组件
  // 'modules/ImageDrop': ImageDrop, //复制粘贴组件
}, true);

interface SuggestFormInputProps {
  value?: string;
  onChange?: (value: string) => void;
}

const SuggestFormInput: React.FC<SuggestFormInputProps> = ({  value, onChange }) => {
  const [content, setContent] = useState<string>( "");
  
  const curChange = (val: string) => {
    onChange?.(val);
  };

  useEffect(() => {
    if(value) {
      setContent(value)
    }
  }, [value])

  const modules = {
    toolbar: [
      ['bold', 'italic', 'underline', 'strike'],        // toggled buttons
      ['blockquote', 'code-block'],
      ['link', 'image'],

      [{ 'header': 1 }, { 'header': 2 }],               // custom button values
      [{ 'list': 'ordered' }, { 'list': 'bullet' }],
      [{ 'script': 'sub' }, { 'script': 'super' }],      // superscript/subscript
      [{ 'indent': '-1' }, { 'indent': '+1' }],          // outdent/indent
      [{ 'direction': 'rtl' }],                         // text direction

      // [{ 'size': ['small', false, 'large', 'huge'] }],  // custom dropdown
      [{ 'header': [1, 2, 3, 4, 5, 6, false] }],

      [{ 'color': [] }, { 'background': [] }],          // dropdown with defaults from theme
      [{ 'font': [] }],
      [{ 'align': [] }],

      ['clean']                                         // remove formatting button
    ]
  }

  return <ReactQuill
    id={"content"}
    key="content"
    defaultValue={value || content}
    value={value || content}
    theme="snow"
    modules={modules}
    // formats={this.formats}
    className={"ql-editor content"}
    onChange={(val) => curChange(val)}
  />
}

export default SuggestFormInput;
