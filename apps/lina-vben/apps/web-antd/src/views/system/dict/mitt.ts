import mitt from 'mitt';

type Events = {
  rowClick: string;
};

export const emitter = mitt<Events>();
