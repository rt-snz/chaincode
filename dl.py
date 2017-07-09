# coding: utf-8
import sys, os
sys.path.append(os.pardir)  # 親ディレクトリのファイルをインポートするための設定
import numpy as np
import pickle
from dataset.mnist import load_mnist

def sigmoid(x):
    return 1 / (1 + np.exp(-x))

def softmax(a):
    c = np.max(a)
    exp_a = np.exp(a - c)
    sum_exp_a = np.sum(exp_a)
    y = exp_a / sum_exp_a
    return y

def get_data():
    (train_img, train_lbl), (test_img, test_lbl) = load_mnist(normalize=True, flatten=True, one_hot_label=False)
    return test_img, test_lbl

def init_network():
    with open("sample_weight.pkl", 'rb') as f:
        network = pickle.load(f)
    return network

def predict(network, x):
    W1, W2, W3 = network['W1'], network['W2'], network['W3']
    b1, b2, b3 = network['b1'], network['b2'], network['b3']

    a1 = np.dot(x, W1) + b1
    z1 = sigmoid(a1)
    a2 = np.dot(z1, W2) + b2
    z2 = sigmoid(a2)
    a3 = np.dot(z2, W3) + b3
    y = softmax(a3)

    return y

def cross_entropy_error(y, t):
    if y.ndim == 1:
        t = t.reshape(1, t.size)
        y = y.reshape(1, y.size)

    batch_size = y.shape[0]
    return -np.sum(t * np.log(y + delta))

###
# main
###

img, lbl = get_data()
network = init_network()

batch_size = 100
accuracy_cnt = 0

print("len(img) = " + str(len(img)))
x_batch = img[0:100]
print("x_batch.shape = " + str(x_batch.shape))
#for i in range(len(img)):
for i in range(0, len(img), batch_size):
    x_batch = img[i:i+batch_size]
    y_batch = predict(network, x_batch)
    print(y_batch.shape)
    p = np.argmax(y_batch, axis=1) # 最も確率の高い要素のインデックスを取得
    accuracy_cnt += np.sum(p == lbl[i:i+batch_size])

print("Accuracy:" + str(float(accuracy_cnt) / len(img)))
