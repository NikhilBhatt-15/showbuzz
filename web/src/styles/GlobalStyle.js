import { createGlobalStyle } from 'styled-components';

export const GlobalStyle = createGlobalStyle`
  *, *::before, *::after { box-sizing: border-box; }
  html, body { margin: 0; padding: 0; }
  body {
    font-family: Roboto, Arial, sans-serif;
    color: ${(p) => p.theme.text};
    background: #fff;
    -webkit-font-smoothing: antialiased;
    font-size: 14px;
  }
  a { color: inherit; text-decoration: none; }
  button { font-family: inherit; cursor: pointer; }
  button:disabled { cursor: not-allowed; }
  input, select, textarea { font-family: inherit; }
  img { display: block; max-width: 100%; }
  h1, h2, h3, h4, p { margin: 0; }
  ::selection { background: ${(p) => p.theme.brandSoft}; }
`;
